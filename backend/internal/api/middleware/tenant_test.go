package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/api/middleware"
)

func TestTenantContext_ExtractsTenantID(t *testing.T) {
	tenantID := uuid.New()

	handler := middleware.TenantContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := middleware.GetTenantUUID(r.Context())
		if !ok {
			t.Error("Expected tenant ID in context")
		}
		if id != tenantID {
			t.Errorf("Expected %s, got %s", tenantID, id)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestTenantContext_RejectsInvalidUUID(t *testing.T) {
	handler := middleware.TenantContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Tenant-ID", "invalid-uuid")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestRequireTenant_RejectsMissingTenant(t *testing.T) {
	handler := middleware.TenantContext(
		middleware.RequireTenant(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler should not be called")
		})),
	)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestTenantContext_ExtractsOperatorID(t *testing.T) {
	operatorID := uuid.New()

	handler := middleware.TenantContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := middleware.GetOperatorUUID(r.Context())
		if !ok {
			t.Error("Expected operator ID in context")
		}
		if id != operatorID {
			t.Errorf("Expected %s, got %s", operatorID, id)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Operator-ID", operatorID.String())
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}
