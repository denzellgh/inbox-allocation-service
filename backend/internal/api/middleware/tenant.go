package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/api/response"
	"github.com/inbox-allocation-service/internal/pkg/logger"
)

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey = logger.TenantIDKey
	// OperatorIDKey is the context key for operator ID
	OperatorIDKey = logger.OperatorIDKey

	// Header names
	TenantIDHeader   = "X-Tenant-ID"
	OperatorIDHeader = "X-Operator-ID"
)

// TenantContext middleware extracts tenant and operator IDs from headers
// Required: Enforces multi-tenancy at HTTP layer
func TenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract Tenant ID (required for most endpoints)
		tenantIDStr := r.Header.Get(TenantIDHeader)
		if tenantIDStr != "" {
			tenantID, err := uuid.Parse(tenantIDStr)
			if err != nil {
				response.BadRequest(w, "Invalid tenant ID format")
				return
			}
			ctx = context.WithValue(ctx, TenantIDKey, tenantID)
		}

		// Extract Operator ID (optional, required for operator-specific actions)
		operatorIDStr := r.Header.Get(OperatorIDHeader)
		if operatorIDStr != "" {
			operatorID, err := uuid.Parse(operatorIDStr)
			if err != nil {
				response.BadRequest(w, "Invalid operator ID format")
				return
			}
			ctx = context.WithValue(ctx, OperatorIDKey, operatorID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireTenant middleware ensures tenant ID is present
func RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Value(TenantIDKey).(uuid.UUID); !ok {
			response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired,
				"X-Tenant-ID header is required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireOperator middleware ensures operator ID is present
func RequireOperator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Value(OperatorIDKey).(uuid.UUID); !ok {
			response.Error(w, http.StatusBadRequest, response.ErrCodeOperatorRequired,
				"X-Operator-ID header is required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetTenantID extracts tenant ID from context
func GetTenantID(ctx context.Context) string {
	if id, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok {
		return id.String()
	}
	return ""
}

// GetTenantUUID extracts tenant ID as UUID from context
func GetTenantUUID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	return id, ok
}

// GetOperatorID extracts operator ID from context
func GetOperatorID(ctx context.Context) string {
	if id, ok := ctx.Value(OperatorIDKey).(uuid.UUID); ok {
		return id.String()
	}
	return ""
}

// GetOperatorUUID extracts operator ID as UUID from context
func GetOperatorUUID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(OperatorIDKey).(uuid.UUID)
	return id, ok
}
