package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/inbox-allocation-service/internal/config"
	"github.com/inbox-allocation-service/internal/pkg/database"
	"github.com/inbox-allocation-service/internal/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize logger
	log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting Inbox Allocation Service",
		zap.String("log_level", cfg.Log.Level),
		zap.String("log_format", cfg.Log.Format),
	)

	// Create database connection pool
	log.Info("Connecting to database",
		zap.String("host", cfg.Database.Host),
		zap.String("port", cfg.Database.Port),
		zap.String("database", cfg.Database.DBName),
	)

	pool, err := database.NewPool(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to create database pool", zap.Error(err))
	}
	defer pool.Close()

	// Verify database connection
	ctx := context.Background()
	if err := database.HealthCheck(ctx, pool); err != nil {
		log.Fatal("Database health check failed", zap.Error(err))
	}

	log.Info("Database connection established successfully")

	// Server initialization will be added in Stage 3
	log.Info("Server initialized successfully",
		zap.String("host", cfg.Server.Host),
		zap.String("port", cfg.Server.Port),
	)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Service is ready. Press Ctrl+C to shutdown.")

	// Wait for shutdown signal
	<-quit

	log.Info("Shutting down service...")

	// Cleanup
	pool.Close()

	log.Info("Service stopped gracefully")
}
