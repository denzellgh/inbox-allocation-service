package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/inbox-allocation-service/internal/pkg/logger"
	"go.uber.org/zap"
)

// Logger returns a middleware that logs HTTP requests
func Logger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create request-scoped logger with context fields
			reqLogger := log.WithContext(r.Context()).WithFields(
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Int64("content_length", r.ContentLength),
			)

			// Log request start at debug level
			reqLogger.Debug("request started")

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Add logger to context for downstream use
			ctx := logger.WithLogger(r.Context(), reqLogger)

			// Call next handler
			next.ServeHTTP(ww, r.WithContext(ctx))

			// Calculate duration
			duration := time.Since(start)

			// Log based on status code
			fields := []zap.Field{
				zap.Int("status", ww.Status()),
				zap.Int("bytes", ww.BytesWritten()),
				zap.Duration("duration", duration),
				zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
			}

			switch {
			case ww.Status() >= 500:
				reqLogger.Error("request completed with server error", fields...)
			case ww.Status() >= 400:
				reqLogger.Warn("request completed with client error", fields...)
			default:
				reqLogger.Info("request completed", fields...)
			}
		})
	}
}

// LoggerWithSampling returns a logger that samples high-volume requests
func LoggerWithSampling(log *logger.Logger, sampleRate int) func(http.Handler) http.Handler {
	counter := 0
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			counter++
			shouldLog := counter%sampleRate == 0

			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			reqLogger := log.WithContext(r.Context())
			ctx := logger.WithLogger(r.Context(), reqLogger)

			next.ServeHTTP(ww, r.WithContext(ctx))

			if shouldLog || ww.Status() >= 400 {
				reqLogger.Info("request completed",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", ww.Status()),
					zap.Duration("duration", time.Since(start)),
				)
			}
		})
	}
}
