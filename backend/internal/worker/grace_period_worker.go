package worker

import (
	"context"
	"sync"
	"time"

	"github.com/inbox-allocation-service/internal/pkg/logger"
	"github.com/inbox-allocation-service/internal/service"
	"go.uber.org/zap"
)

// GracePeriodWorkerConfig holds configuration for the grace period worker
type GracePeriodWorkerConfig struct {
	Interval  time.Duration
	BatchSize int
}

// DefaultGracePeriodWorkerConfig returns sensible defaults
func DefaultGracePeriodWorkerConfig() GracePeriodWorkerConfig {
	return GracePeriodWorkerConfig{
		Interval:  30 * time.Second,
		BatchSize: 100,
	}
}

// GracePeriodWorker processes expired grace periods
type GracePeriodWorker struct {
	service *service.GracePeriodService
	config  GracePeriodWorkerConfig
	logger  *logger.Logger

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewGracePeriodWorker creates a new grace period worker
func NewGracePeriodWorker(
	svc *service.GracePeriodService,
	config GracePeriodWorkerConfig,
	log *logger.Logger,
) *GracePeriodWorker {
	return &GracePeriodWorker{
		service: svc,
		config:  config,
		logger:  log,
		stopCh:  make(chan struct{}),
	}
}

// Name returns the worker's name
func (w *GracePeriodWorker) Name() string {
	return "GracePeriodWorker"
}

// Start begins the worker's processing loop
func (w *GracePeriodWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	defer w.wg.Done()

	w.logger.Info("Grace period worker started",
		zap.Duration("interval", w.config.Interval),
		zap.Int("batch_size", w.config.BatchSize))

	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

	// Process immediately on start
	w.process(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Grace period worker stopping due to context cancellation")
			return
		case <-w.stopCh:
			w.logger.Info("Grace period worker stopping due to stop signal")
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

// Stop gracefully stops the worker
func (w *GracePeriodWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	w.logger.Info("Grace period worker stopped")
}

// process runs a single processing cycle
func (w *GracePeriodWorker) process(ctx context.Context) {
	start := time.Now()

	result, err := w.service.ProcessExpiredGracePeriods(ctx, w.config.BatchSize)
	if err != nil {
		w.logger.Error("Failed to process grace periods",
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return
	}

	// Only log if there was activity
	if result.Processed > 0 {
		w.logger.Info("Grace period worker cycle completed",
			zap.Int("processed", result.Processed),
			zap.Int("transitioned", result.Transitioned),
			zap.Int("already_handled", result.AlreadyHandled),
			zap.Int("errors", result.Errors),
			zap.Duration("duration", time.Since(start)))
	} else {
		w.logger.Debug("Grace period worker cycle completed - no expired periods")
	}
}
