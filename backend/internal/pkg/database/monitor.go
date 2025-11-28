package database

import (
	"context"
	"time"

	"github.com/inbox-allocation-service/internal/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PoolStats represents connection pool statistics
type PoolStats struct {
	TotalConns      int32
	IdleConns       int32
	AcquiredConns   int32
	MaxConns        int32
	AcquireCount    int64
	AcquireDuration time.Duration
}

// GetPoolStats returns current pool statistics
func GetPoolStats(pool *pgxpool.Pool) PoolStats {
	stat := pool.Stat()
	return PoolStats{
		TotalConns:      stat.TotalConns(),
		IdleConns:       stat.IdleConns(),
		AcquiredConns:   stat.AcquiredConns(),
		MaxConns:        stat.MaxConns(),
		AcquireCount:    stat.AcquireCount(),
		AcquireDuration: stat.AcquireDuration(),
	}
}

// StartPoolMonitor starts periodic pool stats logging
func StartPoolMonitor(ctx context.Context, pool *pgxpool.Pool, log *logger.Logger, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Info("starting connection pool monitor",
		zap.Duration("interval", interval),
	)

	for {
		select {
		case <-ctx.Done():
			log.Info("stopping connection pool monitor")
			return
		case <-ticker.C:
			stats := GetPoolStats(pool)
			log.Debug("connection pool stats",
				zap.Int32("total", stats.TotalConns),
				zap.Int32("idle", stats.IdleConns),
				zap.Int32("acquired", stats.AcquiredConns),
				zap.Int32("max", stats.MaxConns),
				zap.Int64("acquire_count", stats.AcquireCount),
				zap.Duration("acquire_duration", stats.AcquireDuration),
			)

			// Warn if pool is near capacity
			if float64(stats.AcquiredConns)/float64(stats.MaxConns) > 0.8 {
				log.Warn("connection pool near capacity",
					zap.Int32("acquired", stats.AcquiredConns),
					zap.Int32("max", stats.MaxConns),
					zap.Float64("usage_percent", float64(stats.AcquiredConns)/float64(stats.MaxConns)*100),
				)
			}
		}
	}
}
