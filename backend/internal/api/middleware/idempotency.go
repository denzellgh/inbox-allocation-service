package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/inbox-allocation-service/internal/service"
)

const (
	// IdempotencyKeyHeader is the header name for idempotency keys
	IdempotencyKeyHeader = "X-Idempotency-Key"
	// IdempotencyReplayHeader indicates a replayed response
	IdempotencyReplayHeader = "X-Idempotency-Replay"
)

// responseRecorder captures the response for caching
type responseRecorder struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Idempotency creates middleware for idempotency key handling
func Idempotency(svc *service.IdempotencyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to mutation methods
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
				next.ServeHTTP(w, r)
				return
			}

			// Get idempotency key from header
			key := r.Header.Get(IdempotencyKeyHeader)
			if key == "" {
				// No idempotency key provided, proceed normally
				next.ServeHTTP(w, r)
				return
			}

			// Get tenant ID from context
			tenantID, ok := GetTenantUUID(r.Context())
			if !ok {
				// No tenant, proceed without idempotency
				next.ServeHTTP(w, r)
				return
			}

			// Read request body for hashing
			var requestBody []byte
			if r.Body != nil {
				requestBody, _ = io.ReadAll(r.Body)
				r.Body.Close()
				// Restore body for handler
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			// Check if key exists
			cached, err := svc.CheckKey(r.Context(), tenantID, key, requestBody)
			if err != nil {
				if err == service.ErrRequestHashMismatch {
					http.Error(w, "Idempotency key reused with different request", http.StatusUnprocessableEntity)
					return
				}
				// Log error but proceed with request
				next.ServeHTTP(w, r)
				return
			}

			// If cached response exists, return it
			if cached != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Header().Set(IdempotencyReplayHeader, "true")
				w.WriteHeader(cached.Status)
				w.Write(cached.Body)
				return
			}

			// No cached response, execute request and capture result
			recorder := newResponseRecorder(w)
			next.ServeHTTP(recorder, r)

			// Store result (only for successful responses or specific errors)
			// Store for 2xx and 4xx (not 5xx which might be transient)
			if recorder.status < 500 {
				err := svc.StoreResult(
					r.Context(),
					tenantID,
					key,
					r.URL.Path,
					r.Method,
					requestBody,
					recorder.status,
					recorder.body.Bytes(),
				)
				if err != nil {
					// Log error but don't fail the request
					// The response was already sent
				}
			}
		})
	}
}
