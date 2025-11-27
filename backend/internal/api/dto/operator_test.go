package dto_test

import (
	"testing"

	"github.com/inbox-allocation-service/internal/api/dto"
)

func TestUpdateStatusRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"valid AVAILABLE", "AVAILABLE", false},
		{"valid OFFLINE", "OFFLINE", false},
		{"invalid status", "INVALID", true},
		{"empty status", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dto.UpdateStatusRequest{Status: tt.status}
			errs := req.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation error")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected validation errors: %v", errs)
			}
		})
	}
}

func TestCreateOperatorRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{"valid OPERATOR", "OPERATOR", false},
		{"valid MANAGER", "MANAGER", false},
		{"valid ADMIN", "ADMIN", false},
		{"invalid role", "SUPERUSER", true},
		{"empty role", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dto.CreateOperatorRequest{Role: tt.role}
			errs := req.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation error")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
		})
	}
}

func TestUpdateOperatorRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{"valid OPERATOR", "OPERATOR", false},
		{"valid MANAGER", "MANAGER", false},
		{"valid ADMIN", "ADMIN", false},
		{"invalid role", "GUEST", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dto.UpdateOperatorRequest{Role: tt.role}
			errs := req.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation error")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
		})
	}
}
