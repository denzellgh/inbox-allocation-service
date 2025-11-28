package database

import (
	"context"
	"fmt"
	"time"

	"github.com/inbox-allocation-service/internal/config"
	"github.com/inbox-allocation-service/internal/pkg/logger"
	"github.com/inbox-allocation-service/internal/pkg/retry"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// NewPool creates a new PostgreSQL connection pool
func NewPool(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	// Configure pool
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MinConns = int32(cfg.MinConns)
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30
	poolConfig.HealthCheckPeriod = time.Minute

	// Create pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	return pool, nil
}

// NewPoolWithRetry creates a pool with connection retry
func NewPoolWithRetry(cfg *config.DatabaseConfig, log *logger.Logger) (*pgxpool.Pool, error) {
	retryCfg := retry.Config{
		MaxAttempts:    5,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
		Jitter:         0.1,
		OnRetry: func(attempt int, err error, nextBackoff time.Duration) {
			log.Warn("database connection attempt failed",
				zap.Int("attempt", attempt),
				zap.Error(err),
				zap.Duration("next_retry_in", nextBackoff),
			)
		},
	}

	pool, err := retry.DoWithResult(context.Background(), retryCfg, func() (*pgxpool.Pool, error) {
		pool, err := NewPool(cfg)
		if err != nil {
			return nil, err
		}

		// Verify connection works
		if err := HealthCheck(context.Background(), pool); err != nil {
			pool.Close()
			return nil, err
		}

		return pool, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect after retries: %w", err)
	}

	log.Info("database connection established",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.DBName),
	)

	return pool, nil
}

// HealthCheck verifies database connectivity
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result int
	err := pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected health check result: %d", result)
	}

	return nil
}
