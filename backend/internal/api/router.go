package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/inbox-allocation-service/internal/api/handler"
	"github.com/inbox-allocation-service/internal/api/middleware"
	"github.com/inbox-allocation-service/internal/pkg/logger"
	"github.com/inbox-allocation-service/internal/repository"
	"github.com/inbox-allocation-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RouterConfig holds dependencies for router creation
type RouterConfig struct {
	Logger             *logger.Logger
	Pool               *pgxpool.Pool
	Repos              *repository.RepositoryContainer
	Services           *ServiceContainer
	IdempotencyService *service.IdempotencyService
	Version            string
	BuildTime          string
	CORSConfig         middleware.CORSConfig
}

// ServiceContainer holds all service instances
type ServiceContainer struct {
	Operator     *service.OperatorService
	Inbox        *service.InboxService
	Subscription *service.SubscriptionService
	Tenant       *service.TenantService
	Conversation *service.ConversationService
	Allocation   *service.AllocationService
	Lifecycle    *service.LifecycleService
	Label        *service.LabelService
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
		// Apply tenant requirement and operator loader to all API routes
		r.Use(middleware.RequireTenant)
		r.Use(middleware.OperatorLoader(cfg.Repos))

		// Initialize handlers
		operatorHandler := handler.NewOperatorHandler(cfg.Services.Operator)
		inboxHandler := handler.NewInboxHandler(cfg.Services.Inbox)
		subscriptionHandler := handler.NewSubscriptionHandler(
			cfg.Services.Subscription,
			cfg.Services.Operator,
			cfg.Services.Inbox,
		)
		tenantHandler := handler.NewTenantHandler(cfg.Services.Tenant)

		// 4.1 Operator Status (any operator)
		r.Route("/operator", func(r chi.Router) {
			r.Use(middleware.RequireOperator)
			r.Get("/status", operatorHandler.GetStatus)
			r.Put("/status", operatorHandler.UpdateStatus)
		})

		// 4.2 & 4.4 Inboxes
		r.Route("/inboxes", func(r chi.Router) {
			r.Get("/", inboxHandler.ListForOperator) // Any operator

			// Admin/Manager only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireManager)
				r.Post("/", inboxHandler.Create)
			})

			r.Route("/{id}", func(r chi.Router) {
				r.Use(middleware.RequireManager)
				r.Get("/", inboxHandler.GetByID)
				r.Put("/", inboxHandler.Update)
				r.Delete("/", inboxHandler.Delete)
			})

			// 4.5 Subscriptions for inbox
			r.Route("/{inbox_id}/operators", func(r chi.Router) {
				r.Use(middleware.RequireManager)
				r.Get("/", subscriptionHandler.ListOperators)
				r.Post("/", subscriptionHandler.Subscribe)
				r.Delete("/{operator_id}", subscriptionHandler.Unsubscribe)
			})
		})

		// 4.3 Operators CRUD (Admin only)
		r.Route("/operators", func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Get("/", operatorHandler.List)
			r.Post("/", operatorHandler.Create)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", operatorHandler.GetByID)
				r.Put("/", operatorHandler.Update)
				r.Delete("/", operatorHandler.Delete)
			})
			// Subscriptions for operator
			r.Get("/{operator_id}/inboxes", subscriptionHandler.ListInboxes)
		})

		// 4.6 Tenant Configuration (Admin only)
		r.Route("/tenant", func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Get("/", tenantHandler.Get)
			r.Put("/weights", tenantHandler.UpdateWeights)
		})

		// 5.1 & 5.2 Conversations (any operator with access)
		conversationHandler := handler.NewConversationHandler(cfg.Services.Conversation)
		r.Route("/conversations", func(r chi.Router) {
			r.Get("/", conversationHandler.List)
			r.Get("/{id}", conversationHandler.GetByID)
		})

		// Search endpoint
		r.Get("/search", conversationHandler.Search)

		// 6.1 & 6.2 Allocation & Claim with Idempotency
		allocationHandler := handler.NewAllocationHandler(cfg.Services.Allocation)
		lifecycleHandler := handler.NewLifecycleHandler(cfg.Services.Lifecycle)

		if cfg.IdempotencyService != nil {
			// Apply idempotency middleware to critical mutation endpoints
			r.Group(func(r chi.Router) {
				r.Use(middleware.Idempotency(cfg.IdempotencyService))
				r.Post("/allocate", allocationHandler.Allocate)
				r.Post("/claim", allocationHandler.Claim)
				r.Post("/resolve", lifecycleHandler.Resolve)
				r.Post("/deallocate", lifecycleHandler.Deallocate)
				r.Post("/reassign", lifecycleHandler.Reassign)
				r.Post("/move_inbox", lifecycleHandler.MoveInbox)
			})
		} else {
			// Without idempotency (fallback)
			r.Post("/allocate", allocationHandler.Allocate)
			r.Post("/claim", allocationHandler.Claim)
			r.Post("/resolve", lifecycleHandler.Resolve)
			r.Post("/deallocate", lifecycleHandler.Deallocate)
			r.Post("/reassign", lifecycleHandler.Reassign)
			r.Post("/move_inbox", lifecycleHandler.MoveInbox)
		}

		// 8.1-8.2 Label Management
		labelHandler := handler.NewLabelHandler(cfg.Services.Label)
		r.Route("/labels", func(r chi.Router) {
			r.Post("/", labelHandler.Create)
			r.Get("/", labelHandler.List)
			r.Put("/{id}", labelHandler.Update)
			r.Delete("/{id}", labelHandler.Delete)

			r.Post("/attach", labelHandler.Attach)
			r.Post("/detach", labelHandler.Detach)
		})
	})

	return r
}
