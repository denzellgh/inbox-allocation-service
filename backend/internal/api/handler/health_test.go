package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/inbox-allocation-service/internal/api/handler"
)

func TestHealthHandler_Version(t *testing.T) {
	h := handler.NewHealthHandler(nil, "1.0.0", "2024-01-01")

	req := httptest.NewRequest("GET", "/version", nil)
	rr := httptest.NewRecorder()
	h.Version(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response["success"].(bool) {
		t.Error("Expected success to be true")
	}

	data := response["data"].(map[string]interface{})
	if data["version"] != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %v", data["version"])
	}

	if data["build_time"] != "2024-01-01" {
		t.Errorf("Expected build_time 2024-01-01, got %v", data["build_time"])
	}
}
