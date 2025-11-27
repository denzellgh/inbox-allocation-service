//go:build integration

package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain sets up the test container
// Note: These tests require Docker to be running

func TestConversationRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pc := testutil.NewPostgresContainer(t)
	ctx := testutil.TestContext(t)

	// Clean tables before each test
	t.Cleanup(func() {
		pc.CleanTables(ctx)
	})

	queries := New(pc.Pool)

	t.Run("create and get conversation", func(t *testing.T) {
		pc.CleanTables(ctx)
		repo := NewConversationRefRepository(queries, pc.Pool)

		// Create tenant first
		tenantRepo := NewTenantRepository(queries)
		tenant := testutil.NewTestTenant()
		err := tenantRepo.Create(ctx, tenant)
		require.NoError(t, err)

		// Create inbox
		inboxRepo := NewInboxRepository(queries)
		inbox := testutil.NewTestInbox(tenant.ID)
		err = inboxRepo.Create(ctx, inbox)
		require.NoError(t, err)

		// Create conversation
		conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
		err = repo.Create(ctx, conv)
		require.NoError(t, err)

		// Get conversation
		retrieved, err := repo.GetByID(ctx, conv.ID)
		require.NoError(t, err)
		assert.Equal(t, conv.ID, retrieved.ID)
		assert.Equal(t, conv.ExternalConversationID, retrieved.ExternalConversationID)
		assert.Equal(t, domain.ConversationStateQueued, retrieved.State)
	})

	t.Run("update conversation state", func(t *testing.T) {
		pc.CleanTables(ctx)
		repo := NewConversationRefRepository(queries, pc.Pool)

		// Setup
		tenantRepo := NewTenantRepository(queries)
		tenant := testutil.NewTestTenant()
		tenantRepo.Create(ctx, tenant)

		inboxRepo := NewInboxRepository(queries)
		inbox := testutil.NewTestInbox(tenant.ID)
		inboxRepo.Create(ctx, inbox)

		operatorRepo := NewOperatorRepository(queries)
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)
		operatorRepo.Create(ctx, operator)

		// Create conversation
		conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
		repo.Create(ctx, conv)

		// Update state
		err := conv.Allocate(operator.ID)
		require.NoError(t, err)
		err = repo.Update(ctx, conv)
		require.NoError(t, err)

		// Verify
		retrieved, _ := repo.GetByID(ctx, conv.ID)
		assert.Equal(t, domain.ConversationStateAllocated, retrieved.State)
		assert.Equal(t, operator.ID, *retrieved.AssignedOperatorID)
	})

	t.Run("get next for allocation with lock", func(t *testing.T) {
		pc.CleanTables(ctx)
		repo := NewConversationRefRepository(queries, pc.Pool)

		// Setup
		tenantRepo := NewTenantRepository(queries)
		tenant := testutil.NewTestTenant()
		tenantRepo.Create(ctx, tenant)

		inboxRepo := NewInboxRepository(queries)
		inbox := testutil.NewTestInbox(tenant.ID)
		inboxRepo.Create(ctx, inbox)

		// Create queued conversations
		for i := 0; i < 5; i++ {
			conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
			repo.Create(ctx, conv)
		}

		// Get next for allocation (uses FOR UPDATE SKIP LOCKED)
		convs, err := repo.GetNextForAllocation(ctx, tenant.ID, []uuid.UUID{inbox.ID}, 3)
		require.NoError(t, err)
		assert.Len(t, convs, 3)

		// All should be QUEUED
		for _, conv := range convs {
			assert.Equal(t, domain.ConversationStateQueued, conv.State)
		}
	})
}

func TestIdempotencyRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pc := testutil.NewPostgresContainer(t)
	ctx := testutil.TestContext(t)

	queries := New(pc.Pool)

	t.Run("create and get idempotency key", func(t *testing.T) {
		pc.CleanTables(ctx)
		repo := NewIdempotencyRepository(queries)

		// Setup tenant
		tenantRepo := NewTenantRepository(queries)
		tenant := testutil.NewTestTenant()
		tenantRepo.Create(ctx, tenant)

		// Create idempotency key
		ik := domain.NewIdempotencyKey(
			"test-key",
			tenant.ID,
			"/api/v1/allocate",
			"POST",
			nil,
			200,
			[]byte(`{"success": true}`),
			24*time.Hour,
		)
		err := repo.Create(ctx, ik)
		require.NoError(t, err)

		// Get by key
		retrieved, err := repo.GetByKey(ctx, tenant.ID, "test-key")
		require.NoError(t, err)
		assert.Equal(t, ik.ID, retrieved.ID)
		assert.Equal(t, 200, retrieved.ResponseStatus)
	})

	t.Run("unique constraint on tenant + key", func(t *testing.T) {
		pc.CleanTables(ctx)
		repo := NewIdempotencyRepository(queries)

		// Setup tenant
		tenantRepo := NewTenantRepository(queries)
		tenant := testutil.NewTestTenant()
		tenantRepo.Create(ctx, tenant)

		// Create first key
		ik1 := domain.NewIdempotencyKey(
			"duplicate-key",
			tenant.ID,
			"/api",
			"POST",
			nil,
			200,
			[]byte("{}"),
			24*time.Hour,
		)
		err := repo.Create(ctx, ik1)
		require.NoError(t, err)

		// Create duplicate - should fail
		ik2 := domain.NewIdempotencyKey(
			"duplicate-key",
			tenant.ID,
			"/api",
			"POST",
			nil,
			200,
			[]byte("{}"),
			24*time.Hour,
		)
		err = repo.Create(ctx, ik2)
		assert.Error(t, err)
	})

	t.Run("delete expired keys", func(t *testing.T) {
		pc.CleanTables(ctx)
		repo := NewIdempotencyRepository(queries)

		// Setup tenant
		tenantRepo := NewTenantRepository(queries)
		tenant := testutil.NewTestTenant()
		tenantRepo.Create(ctx, tenant)

		// Create expired key
		ik := domain.NewIdempotencyKey(
			"expired-key",
			tenant.ID,
			"/api",
			"POST",
			nil,
			200,
			[]byte("{}"),
			-1*time.Hour, // Expired
		)
		repo.Create(ctx, ik)

		// Delete expired
		count, err := repo.DeleteExpired(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Verify deleted
		_, err = repo.GetByKey(ctx, tenant.ID, "expired-key")
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestOperatorStatusRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pc := testutil.NewPostgresContainer(t)
	ctx := testutil.TestContext(t)

	queries := New(pc.Pool)

	t.Run("create and update operator status", func(t *testing.T) {
		pc.CleanTables(ctx)
		repo := NewOperatorStatusRepository(queries)

		// Setup
		tenantRepo := NewTenantRepository(queries)
		tenant := testutil.NewTestTenant()
		tenantRepo.Create(ctx, tenant)

		operatorRepo := NewOperatorRepository(queries)
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)
		operatorRepo.Create(ctx, operator)

		// Create status
		status := testutil.NewTestOperatorStatus(operator.ID, domain.OperatorStatusOffline)
		err := repo.Create(ctx, status)
		require.NoError(t, err)

		// Verify created
		retrieved, err := repo.GetByOperatorID(ctx, operator.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.OperatorStatusOffline, retrieved.Status)

		// Update status
		status.SetStatus(domain.OperatorStatusAvailable)
		err = repo.Update(ctx, status)
		require.NoError(t, err)

		// Verify updated
		retrieved, err = repo.GetByOperatorID(ctx, operator.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.OperatorStatusAvailable, retrieved.Status)
	})
}

func TestGracePeriodRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pc := testutil.NewPostgresContainer(t)
	ctx := testutil.TestContext(t)

	queries := New(pc.Pool)

	t.Run("create and get grace period", func(t *testing.T) {
		pc.CleanTables(ctx)
		repo := NewGracePeriodRepository(queries, pc.Pool)

		// Setup
		tenantRepo := NewTenantRepository(queries)
		tenant := testutil.NewTestTenant()
		tenantRepo.Create(ctx, tenant)

		inboxRepo := NewInboxRepository(queries)
		inbox := testutil.NewTestInbox(tenant.ID)
		inboxRepo.Create(ctx, inbox)

		operatorRepo := NewOperatorRepository(queries)
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)
		operatorRepo.Create(ctx, operator)

		convRepo := NewConversationRefRepository(queries, pc.Pool)
		conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
		convRepo.Create(ctx, conv)

		// Create grace period
		gpa := testutil.NewTestGracePeriod(
			conv.ID,
			operator.ID,
			time.Now().UTC().Add(5*time.Minute),
		)
		err := repo.Create(ctx, gpa)
		require.NoError(t, err)

		// Get by operator
		gpas, err := repo.GetByOperatorID(ctx, operator.ID)
		require.NoError(t, err)
		assert.Len(t, gpas, 1)
		assert.Equal(t, conv.ID, gpas[0].ConversationID)
	})

	t.Run("delete expired grace periods", func(t *testing.T) {
		pc.CleanTables(ctx)
		repo := NewGracePeriodRepository(queries, pc.Pool)

		// Setup
		tenantRepo := NewTenantRepository(queries)
		tenant := testutil.NewTestTenant()
		tenantRepo.Create(ctx, tenant)

		inboxRepo := NewInboxRepository(queries)
		inbox := testutil.NewTestInbox(tenant.ID)
		inboxRepo.Create(ctx, inbox)

		operatorRepo := NewOperatorRepository(queries)
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)
		operatorRepo.Create(ctx, operator)

		convRepo := NewConversationRefRepository(queries, pc.Pool)
		conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
		convRepo.Create(ctx, conv)

		// Create expired grace period
		gpa := testutil.NewTestGracePeriod(
			conv.ID,
			operator.ID,
			time.Now().UTC().Add(-5*time.Minute), // Already expired
		)
		repo.Create(ctx, gpa)

		// Get and lock expired
		expired, err := repo.GetAndLockExpired(ctx, 10)
		require.NoError(t, err)
		assert.Len(t, expired, 1)

		// Delete
		err = repo.Delete(ctx, expired[0].ID)
		require.NoError(t, err)

		// Verify deleted
		remaining, _ := repo.GetByOperatorID(ctx, operator.ID)
		assert.Len(t, remaining, 0)
	})
}
