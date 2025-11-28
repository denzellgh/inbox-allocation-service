package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

// ==================== Status ====================

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

func (r *UpdateStatusRequest) Validate() []string {
	var errs []string
	status := domain.OperatorStatusType(r.Status)
	if !status.IsValid() {
		errs = append(errs, "status must be AVAILABLE or OFFLINE")
	}
	return errs
}

type OperatorStatusResponse struct {
	OperatorID         uuid.UUID `json:"operator_id"`
	Status             string    `json:"status"`
	LastStatusChangeAt time.Time `json:"last_status_change_at"`
}

// ==================== CRUD ====================

type CreateOperatorRequest struct {
	Role string `json:"role"`
}

func (r *CreateOperatorRequest) Validate() []string {
	var errs []string
	role := domain.OperatorRole(r.Role)
	if !role.IsValid() {
		errs = append(errs, "role must be OPERATOR, MANAGER, or ADMIN")
	}
	return errs
}

type UpdateOperatorRequest struct {
	Role string `json:"role"`
}

func (r *UpdateOperatorRequest) Validate() []string {
	var errs []string
	role := domain.OperatorRole(r.Role)
	if !role.IsValid() {
		errs = append(errs, "role must be OPERATOR, MANAGER, or ADMIN")
	}
	return errs
}

type OperatorResponse struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewOperatorResponse(op *domain.Operator) OperatorResponse {
	return OperatorResponse{
		ID:        op.ID,
		TenantID:  op.TenantID,
		Role:      string(op.Role),
		CreatedAt: op.CreatedAt,
		UpdatedAt: op.UpdatedAt,
	}
}

type OperatorListResponse struct {
	Operators []OperatorResponse `json:"operators"`
	Meta      ListMeta           `json:"meta"`
}
