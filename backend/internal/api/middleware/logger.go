package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	size        int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// Logger creates a logging middleware with the provided logger
func Logger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := wrapResponseWriter(w)

			defer func() {
				requestID := GetRequestID(r.Context())
				tenantID := GetTenantID(r.Context())
				operatorID := GetOperatorID(r.Context())

				fields := []zap.Field{
					zap.String("request_id", requestID),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("query", r.URL.RawQuery),
					zap.Int("status", wrapped.status),
					zap.Int("size", wrapped.size),
					zap.Duration("duration", time.Since(start)),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("user_agent", r.UserAgent()),
				}

				if tenantID != "" {
					fields = append(fields, zap.String("tenant_id", tenantID))
				}
				if operatorID != "" {
					fields = append(fields, zap.String("operator_id", operatorID))
				}

				// Log level based on status code
				switch {
				case wrapped.status >= 500:
					log.Error("HTTP request", fields...)
				case wrapped.status >= 400:
					log.Warn("HTTP request", fields...)
				default:
					log.Info("HTTP request", fields...)
				}
			}()

			next.ServeHTTP(wrapped, r)
		})
	}
}
