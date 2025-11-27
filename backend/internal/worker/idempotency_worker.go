package worker

import (
	"context"
	"sync"
	"time"

	"github.com/inbox-allocation-service/internal/service"
	"go.uber.org/zap"
)

// IdempotencyWorkerConfig holds configuration for the idempotency cleanup worker
type IdempotencyWorkerConfig struct {
	Interval time.Duration
}

// DefaultIdempotencyWorkerConfig returns sensible defaults
func DefaultIdempotencyWorkerConfig() IdempotencyWorkerConfig {
	return IdempotencyWorkerConfig{
		Interval: 1 * time.Hour,
	}
}

// IdempotencyWorker cleans up expired idempotency keys
type IdempotencyWorker struct {
	service *service.IdempotencyService
	config  IdempotencyWorkerConfig
	logger  *zap.Logger

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewIdempotencyWorker creates a new idempotency cleanup worker
func NewIdempotencyWorker(
	svc *service.IdempotencyService,
	config IdempotencyWorkerConfig,
	logger *zap.Logger,
) *IdempotencyWorker {
	return &IdempotencyWorker{
		service: svc,
		config:  config,
		logger:  logger,
		stopCh:  make(chan struct{}),
	}
}

// Name returns the worker's name
func (w *IdempotencyWorker) Name() string {
	return "IdempotencyCleanupWorker"
}

// Start begins the worker's processing loop
func (w *IdempotencyWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	defer w.wg.Done()

	w.logger.Info("Idempotency cleanup worker started",
		zap.Duration("interval", w.config.Interval))

	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Idempotency cleanup worker stopping due to context cancellation")
			return
		case <-w.stopCh:
			w.logger.Info("Idempotency cleanup worker stopping due to stop signal")
			return
		case <-ticker.C:
			w.cleanup(ctx)
		}
	}
}

// Stop gracefully stops the worker
func (w *IdempotencyWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	w.logger.Info("Idempotency cleanup worker stopped")
}

// cleanup runs a single cleanup cycle
func (w *IdempotencyWorker) cleanup(ctx context.Context) {
	start := time.Now()

	count, err := w.service.CleanupExpired(ctx)
	if err != nil {
		w.logger.Error("Failed to cleanup expired idempotency keys",
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return
	}

	if count > 0 {
		w.logger.Info("Idempotency cleanup cycle completed",
			zap.Int64("cleaned", count),
			zap.Duration("duration", time.Since(start)))
	}
}
