package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/inbox-allocation-service/internal/api/response"
	"github.com/inbox-allocation-service/internal/pkg/logger"
	"go.uber.org/zap"
)

// Recovery middleware recovers from panics and logs the error
func Recovery(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := logger.GetCorrelationID(r.Context())
					stack := debug.Stack()

					log.Error("Panic recovered",
						zap.String("request_id", requestID),
						zap.Any("error", err),
						zap.String("stack", string(stack)),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)

					response.InternalError(w, "An unexpected error occurred")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
