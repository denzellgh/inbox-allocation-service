package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/shopspring/decimal"
)

type UpdateTenantWeightsRequest struct {
	Alpha float64 `json:"alpha"`
	Beta  float64 `json:"beta"`
}

func (r *UpdateTenantWeightsRequest) Validate() []string {
	var errs []string
	if r.Alpha < 0 || r.Alpha > 1 {
		errs = append(errs, "alpha must be between 0 and 1")
	}
	if r.Beta < 0 || r.Beta > 1 {
		errs = append(errs, "beta must be between 0 and 1")
	}
	sum := r.Alpha + r.Beta
	if sum < 0.99 || sum > 1.01 {
		errs = append(errs, "alpha + beta should equal 1.0")
	}
	return errs
}

func (r *UpdateTenantWeightsRequest) ToDecimal() (alpha, beta decimal.Decimal) {
	return decimal.NewFromFloat(r.Alpha), decimal.NewFromFloat(r.Beta)
}

type TenantResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	PriorityWeightAlpha float64   `json:"priority_weight_alpha"`
	PriorityWeightBeta  float64   `json:"priority_weight_beta"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func NewTenantResponse(t *domain.Tenant) TenantResponse {
	alpha, _ := t.PriorityWeightAlpha.Float64()
	beta, _ := t.PriorityWeightBeta.Float64()
	return TenantResponse{
		ID:                  t.ID,
		Name:                t.Name,
		PriorityWeightAlpha: alpha,
		PriorityWeightBeta:  beta,
		CreatedAt:           t.CreatedAt,
		UpdatedAt:           t.UpdatedAt,
	}
}
