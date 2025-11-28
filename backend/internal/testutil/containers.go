package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps a PostgreSQL testcontainer
type PostgresContainer struct {
	Container testcontainers.Container
	Pool      *pgxpool.Pool
	DSN       string
}

// NewPostgresContainer creates a new PostgreSQL container for testing
func NewPostgresContainer(t *testing.T) *PostgresContainer {
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	// Get connection string
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("Failed to create connection pool: %v", err)
	}

	// Cleanup on test completion
	t.Cleanup(func() {
		pool.Close()
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	})

	pc := &PostgresContainer{
		Container: container,
		Pool:      pool,
		DSN:       dsn,
	}

	// Run migrations
	if err := pc.RunMigrations(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return pc
}

// RunMigrations applies all database migrations
func (pc *PostgresContainer) RunMigrations(ctx context.Context) error {
	// Initial schema migration
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,

		// Tenants
		`CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			priority_weight_alpha DECIMAL(5,4) NOT NULL DEFAULT 0.6,
			priority_weight_beta DECIMAL(5,4) NOT NULL DEFAULT 0.4,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_by UUID
		)`,

		// Inboxes
		`CREATE TABLE IF NOT EXISTS inboxes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			phone_number VARCHAR(20) NOT NULL,
			display_name VARCHAR(255) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(tenant_id, phone_number)
		)`,

		// Operators
		`CREATE TABLE IF NOT EXISTS operators (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			role VARCHAR(20) NOT NULL DEFAULT 'OPERATOR',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		// Operator status
		`CREATE TABLE IF NOT EXISTS operator_status (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			operator_id UUID NOT NULL UNIQUE REFERENCES operators(id) ON DELETE CASCADE,
			status VARCHAR(20) NOT NULL DEFAULT 'OFFLINE',
			last_status_change_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		// Subscriptions
		`CREATE TABLE IF NOT EXISTS operator_inbox_subscriptions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			operator_id UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
			inbox_id UUID NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(operator_id, inbox_id)
		)`,

		// Conversations
		`CREATE TABLE IF NOT EXISTS conversation_refs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			inbox_id UUID NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
			external_conversation_id VARCHAR(255) NOT NULL,
			customer_phone_number VARCHAR(20) NOT NULL,
			state VARCHAR(20) NOT NULL DEFAULT 'QUEUED',
			assigned_operator_id UUID REFERENCES operators(id),
			last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			message_count INT NOT NULL DEFAULT 0,
			priority_score DECIMAL(10,4) NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			resolved_at TIMESTAMPTZ,
			UNIQUE(tenant_id, external_conversation_id)
		)`,

		// Labels
		`CREATE TABLE IF NOT EXISTS labels (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			inbox_id UUID NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			color VARCHAR(7),
			created_by UUID REFERENCES operators(id),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(inbox_id, name)
		)`,

		// Conversation labels
		`CREATE TABLE IF NOT EXISTS conversation_labels (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			conversation_id UUID NOT NULL REFERENCES conversation_refs(id) ON DELETE CASCADE,
			label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(conversation_id, label_id)
		)`,

		// Grace period assignments
		`CREATE TABLE IF NOT EXISTS grace_period_assignments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			conversation_id UUID NOT NULL REFERENCES conversation_refs(id) ON DELETE CASCADE,
			operator_id UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
			expires_at TIMESTAMPTZ NOT NULL,
			reason VARCHAR(20) NOT NULL DEFAULT 'OFFLINE',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		// Idempotency keys
		`CREATE TABLE IF NOT EXISTS idempotency_keys (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			key VARCHAR(255) NOT NULL,
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			endpoint VARCHAR(255) NOT NULL,
			method VARCHAR(10) NOT NULL,
			request_hash VARCHAR(64),
			response_status INT NOT NULL,
			response_body JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL,
			UNIQUE(tenant_id, key)
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_conversation_refs_state ON conversation_refs(state)`,
		`CREATE INDEX IF NOT EXISTS idx_conversation_refs_inbox_state ON conversation_refs(inbox_id, state)`,
		`CREATE INDEX IF NOT EXISTS idx_conversation_refs_priority ON conversation_refs(priority_score DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_grace_period_expires ON grace_period_assignments(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_idempotency_expires ON idempotency_keys(expires_at)`,
	}

	for _, sql := range migrations {
		if _, err := pc.Pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// CleanTables truncates all tables for test isolation
func (pc *PostgresContainer) CleanTables(ctx context.Context) error {
	tables := []string{
		"idempotency_keys",
		"grace_period_assignments",
		"conversation_labels",
		"labels",
		"conversation_refs",
		"operator_inbox_subscriptions",
		"operator_status",
		"operators",
		"inboxes",
		"tenants",
	}

	for _, table := range tables {
		if _, err := pc.Pool.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table)); err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}

	return nil
}
