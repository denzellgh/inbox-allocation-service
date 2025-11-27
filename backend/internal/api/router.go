package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/inbox-allocation-service/internal/api/handler"
	"github.com/inbox-allocation-service/internal/api/middleware"
	"github.com/inbox-allocation-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// RouterConfig holds dependencies for router creation
type RouterConfig struct {
	Logger     *zap.Logger
	Pool       *pgxpool.Pool
	Repos      *repository.RepositoryContainer
	Version    string
	BuildTime  string
	CORSConfig middleware.CORSConfig
}

// NewRouter creates and configures the Chi router
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Global middlewares (order matters!)
	r.Use(middleware.RequestID)            // 1. Request ID first
	r.Use(middleware.CORS(cfg.CORSConfig)) // 2. CORS early
	r.Use(middleware.Recovery(cfg.Logger)) // 3. Recovery before logging
	r.Use(middleware.Logger(cfg.Logger))   // 4. Logging
	r.Use(middleware.TenantContext)        // 5. Tenant context extraction

	// Health check handlers (no tenant required)
	healthHandler := handler.NewHealthHandler(cfg.Pool, cfg.Version, cfg.BuildTime)
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)
	r.Get("/version", healthHandler.Version)

	// API v1 routes (tenant required)
	r.Route("/api/v1", func(r chi.Router) {
		// Apply tenant requirement to all API routes
		r.Use(middleware.RequireTenant)

		// Placeholder for future endpoints
		// r.Route("/operator", ...)
		// r.Route("/inboxes", ...)
		// r.Route("/conversations", ...)
		// r.Route("/labels", ...)
	})

	return r
}
