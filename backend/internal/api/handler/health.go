package handler

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/inbox-allocation-service/internal/api/response"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	pool      *pgxpool.Pool
	version   string
	buildTime string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(pool *pgxpool.Pool, version, buildTime string) *HealthHandler {
	return &HealthHandler{
		pool:      pool,
		version:   version,
		buildTime: buildTime,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string           `json:"status"`
	Checks    map[string]Check `json:"checks"`
	Timestamp time.Time        `json:"timestamp"`
}

// Check represents an individual health check
type Check struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// ReadyResponse represents the readiness response
type ReadyResponse struct {
	Ready     bool      `json:"ready"`
	Timestamp time.Time `json:"timestamp"`
}

// VersionResponse represents the version endpoint response
type VersionResponse struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Health handles GET /health - liveness probe
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]Check)
	overallStatus := "healthy"

	// Check database connectivity
	dbCheck := h.checkDatabase(ctx)
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	healthResponse := HealthResponse{
		Status:    overallStatus,
		Checks:    checks,
		Timestamp: time.Now().UTC(),
	}

	status := http.StatusOK
	if overallStatus != "healthy" {
		status = http.StatusServiceUnavailable
	}

	response.JSON(w, status, healthResponse)
}

// Ready handles GET /ready - readiness probe
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	ready := h.isReady(ctx)

	readyResponse := ReadyResponse{
		Ready:     ready,
		Timestamp: time.Now().UTC(),
	}

	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}

	response.JSON(w, status, readyResponse)
}

// Version handles GET /version - version information
func (h *HealthHandler) Version(w http.ResponseWriter, r *http.Request) {
	versionResponse := VersionResponse{
		Version:   h.version,
		BuildTime: h.buildTime,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	response.OK(w, versionResponse)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) Check {
	err := h.pool.Ping(ctx)
	if err != nil {
		return Check{
			Status:  "unhealthy",
			Message: "Database connection failed",
		}
	}

	// Check pool stats
	stats := h.pool.Stat()
	if stats.TotalConns() == 0 {
		return Check{
			Status:  "unhealthy",
			Message: "No database connections available",
		}
	}

	return Check{
		Status:  "healthy",
		Message: "Connected",
	}
}

func (h *HealthHandler) isReady(ctx context.Context) bool {
	// Check database
	if err := h.pool.Ping(ctx); err != nil {
		return false
	}

	return true
}
