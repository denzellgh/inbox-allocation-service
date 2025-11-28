package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/pkg/logger"
)

const (
	// HeaderXRequestID is the header name for request ID
	HeaderXRequestID = "X-Request-ID"
	// HeaderXCorrelationID is the header name for correlation ID
	HeaderXCorrelationID = "X-Correlation-ID"
)

// RequestID middleware generates or extracts request/correlation IDs
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get existing correlation ID from headers
		correlationID := r.Header.Get(HeaderXCorrelationID)
		if correlationID == "" {
			correlationID = r.Header.Get(HeaderXRequestID)
		}

		// Generate new ID if not provided (using UUIDv7 for time-ordering)
		if correlationID == "" {
			correlationID = uuid.Must(uuid.NewV7()).String()
		}

		// Add to context
		ctx := logger.WithCorrelationIDCtx(r.Context(), correlationID)

		// Add to response headers
		w.Header().Set(HeaderXRequestID, correlationID)
		w.Header().Set(HeaderXCorrelationID, correlationID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts request ID from request context
func GetRequestID(r *http.Request) string {
	return logger.GetCorrelationID(r.Context())
}
