package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/inbox-allocation-service/internal/api/middleware"
)

func TestRequestID_GeneratesID(t *testing.T) {
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := middleware.GetRequestID(r.Context())
		if id == "" {
			t.Error("Expected request ID in context")
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID header in response")
	}
}

func TestRequestID_UsesProvidedID(t *testing.T) {
	providedID := "custom-request-id-123"
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := middleware.GetRequestID(r.Context())
		if id != providedID {
			t.Errorf("Expected %s, got %s", providedID, id)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", providedID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
}
