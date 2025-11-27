package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/inbox-allocation-service/internal/api"
	"github.com/inbox-allocation-service/internal/api/middleware"
	"github.com/inbox-allocation-service/internal/config"
	"github.com/inbox-allocation-service/internal/pkg/database"
	"github.com/inbox-allocation-service/internal/pkg/logger"
	"github.com/inbox-allocation-service/internal/repository"
	"github.com/inbox-allocation-service/internal/server"
	"github.com/inbox-allocation-service/internal/service"
	"go.uber.org/zap"
)

// Build-time variables
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// Initialize logger
	log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting Inbox Allocation Service",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
	)

	// Connect to database
	pool, err := database.NewPool(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Verify database connection
	if err := database.HealthCheck(context.Background(), pool); err != nil {
		log.Fatal("Database health check failed", zap.Error(err))
	}
	log.Info("Database connection established")

	// Initialize repositories
	repos := repository.NewRepositoryContainer(pool)
	log.Info("Repositories initialized")

	// Initialize transaction manager
	txMgr := database.NewTxManager(pool)

	// Initialize services
	services := &api.ServiceContainer{
		Operator:     service.NewOperatorService(repos, txMgr, log),
		Inbox:        service.NewInboxService(repos, log),
		Subscription: service.NewSubscriptionService(repos, log),
		Tenant:       service.NewTenantService(repos, log),
		Conversation: service.NewConversationService(repos, log),
		Allocation:   service.NewAllocationService(repos, pool, log),
	}
	log.Info("Services initialized")

	// Create router
	router := api.NewRouter(api.RouterConfig{
		Logger:     log,
		Pool:       pool,
		Repos:      repos,
		Services:   services,
		Version:    Version,
		BuildTime:  BuildTime,
		CORSConfig: middleware.DefaultCORSConfig(),
	})

	// Parse server port
	port, err := strconv.Atoi(cfg.Server.Port)
	if err != nil {
		log.Fatal("Invalid server port", zap.String("port", cfg.Server.Port), zap.Error(err))
	}

	// Create and start server
	serverConfig := server.Config{
		Host:            cfg.Server.Host,
		Port:            port,
		ReadTimeout:     cfg.Server.ReadTimeout,
		WriteTimeout:    cfg.Server.WriteTimeout,
		IdleTimeout:     cfg.Server.IdleTimeout,
		ShutdownTimeout: cfg.Server.ShutdownTimeout,
	}
	srv := server.New(router, log, serverConfig)

	// Graceful shutdown setup
	go func() {
		if err := srv.Start(); err != nil {
			log.Error("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Received shutdown signal")

	// Shutdown server
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error("Server shutdown error", zap.Error(err))
	}

	log.Info("Service stopped")
}
