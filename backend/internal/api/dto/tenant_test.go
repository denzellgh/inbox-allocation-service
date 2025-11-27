package dto_test

import (
	"testing"

	"github.com/inbox-allocation-service/internal/api/dto"
)

func TestUpdateTenantWeightsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		alpha   float64
		beta    float64
		wantErr bool
	}{
		{"valid 0.5/0.5", 0.5, 0.5, false},
		{"valid 0.6/0.4", 0.6, 0.4, false},
		{"valid 0.7/0.3", 0.7, 0.3, false},
		{"valid 1.0/0.0", 1.0, 0.0, false},
		{"valid 0.0/1.0", 0.0, 1.0, false},
		{"alpha negative", -0.1, 1.1, true},
		{"alpha > 1", 1.5, 0.0, true},
		{"beta negative", 0.5, -0.5, true},
		{"beta > 1", 0.5, 1.5, true},
		{"sum != 1 (too low)", 0.3, 0.3, true},
		{"sum != 1 (too high)", 0.7, 0.7, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dto.UpdateTenantWeightsRequest{Alpha: tt.alpha, Beta: tt.beta}
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

func TestUpdateTenantWeightsRequest_ToDecimal(t *testing.T) {
	req := dto.UpdateTenantWeightsRequest{Alpha: 0.6, Beta: 0.4}
	alpha, beta := req.ToDecimal()

	alphaFloat, _ := alpha.Float64()
	betaFloat, _ := beta.Float64()

	if alphaFloat != 0.6 {
		t.Errorf("alpha: got %v, want 0.6", alphaFloat)
	}
	if betaFloat != 0.4 {
		t.Errorf("beta: got %v, want 0.4", betaFloat)
	}
}
