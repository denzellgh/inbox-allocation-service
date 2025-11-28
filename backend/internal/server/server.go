package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/inbox-allocation-service/internal/pkg/logger"
	"go.uber.org/zap"
)

// ShutdownHook is a function called during shutdown
type ShutdownHook func(ctx context.Context) error

// Config holds server configuration
type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DefaultConfig returns sensible server defaults
func DefaultConfig() Config {
	return Config{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

// Server wraps http.Server with graceful shutdown
type Server struct {
	httpServer *http.Server
	log        *logger.Logger
	config     Config

	mu                sync.Mutex
	preShutdownHooks  []ShutdownHook
	postShutdownHooks []ShutdownHook
}

// New creates a new server instance
func New(handler http.Handler, log *logger.Logger, cfg Config) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		log:    log.Named("server"),
		config: cfg,
	}
}

// OnPreShutdown registers a hook to run before HTTP shutdown
func (s *Server) OnPreShutdown(hook ShutdownHook) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preShutdownHooks = append(s.preShutdownHooks, hook)
}

// OnPostShutdown registers a hook to run after HTTP shutdown
func (s *Server) OnPostShutdown(hook ShutdownHook) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.postShutdownHooks = append(s.postShutdownHooks, hook)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.log.Info("starting HTTP server",
		zap.String("addr", s.httpServer.Addr),
		zap.Duration("read_timeout", s.config.ReadTimeout),
		zap.Duration("write_timeout", s.config.WriteTimeout),
	)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("initiating graceful shutdown",
		zap.Duration("timeout", s.config.ShutdownTimeout),
	)

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	// Run pre-shutdown hooks
	s.log.Debug("running pre-shutdown hooks",
		zap.Int("count", len(s.preShutdownHooks)),
	)
	for i, hook := range s.preShutdownHooks {
		if err := hook(shutdownCtx); err != nil {
			s.log.Warn("pre-shutdown hook failed",
				zap.Int("hook_index", i),
				zap.Error(err),
			)
		}
	}

	// Shutdown HTTP server (waits for in-flight requests)
	s.log.Debug("stopping HTTP server")
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.log.Error("HTTP server shutdown error", zap.Error(err))
		return fmt.Errorf("http shutdown error: %w", err)
	}
	s.log.Info("HTTP server stopped")

	// Run post-shutdown hooks
	s.log.Debug("running post-shutdown hooks",
		zap.Int("count", len(s.postShutdownHooks)),
	)
	for i, hook := range s.postShutdownHooks {
		if err := hook(shutdownCtx); err != nil {
			s.log.Warn("post-shutdown hook failed",
				zap.Int("hook_index", i),
				zap.Error(err),
			)
		}
	}

	s.log.Info("graceful shutdown complete")
	return nil
}

// Addr returns the server address
func (s *Server) Addr() string {
	return s.httpServer.Addr
}
