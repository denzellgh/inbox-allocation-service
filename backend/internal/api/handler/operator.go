package handler

import (
	"net/http"

	"github.com/inbox-allocation-service/internal/api/dto"
	"github.com/inbox-allocation-service/internal/api/middleware"
	"github.com/inbox-allocation-service/internal/api/response"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/service"
)

type OperatorHandler struct {
	service *service.OperatorService
}

func NewOperatorHandler(svc *service.OperatorService) *OperatorHandler {
	return &OperatorHandler{service: svc}
}

// GetStatus handles GET /api/v1/operator/status
func (h *OperatorHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	operatorID, ok := middleware.GetOperatorUUID(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeOperatorRequired, "X-Operator-ID required")
		return
	}

	status, err := h.service.GetStatus(r.Context(), operatorID)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Operator status not found")
			return
		}
		response.InternalError(w, "Failed to get status")
		return
	}

	response.OK(w, dto.OperatorStatusResponse{
		OperatorID:         status.OperatorID,
		Status:             string(status.Status),
		LastStatusChangeAt: status.LastStatusChangeAt,
	})
}

// UpdateStatus handles PUT /api/v1/operator/status
func (h *OperatorHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	operatorID, ok := middleware.GetOperatorUUID(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeOperatorRequired, "X-Operator-ID required")
		return
	}

	req, err := dto.ParseJSON[dto.UpdateStatusRequest](r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	status, err := h.service.UpdateStatus(r.Context(), operatorID, domain.OperatorStatusType(req.Status))
	if err != nil {
		response.InternalError(w, "Failed to update status")
		return
	}

	response.OK(w, dto.OperatorStatusResponse{
		OperatorID:         status.OperatorID,
		Status:             string(status.Status),
		LastStatusChangeAt: status.LastStatusChangeAt,
	})
}

// Create handles POST /api/v1/operators
func (h *OperatorHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantUUID(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired, "X-Tenant-ID required")
		return
	}

	req, err := dto.ParseJSON[dto.CreateOperatorRequest](r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	operator, err := h.service.Create(r.Context(), tenantID, domain.OperatorRole(req.Role))
	if err != nil {
		response.InternalError(w, "Failed to create operator")
		return
	}

	response.Created(w, dto.NewOperatorResponse(operator))
}

// GetByID handles GET /api/v1/operators/{id}
func (h *OperatorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := dto.ParseUUIDParam(r, "id")
	if err != nil {
		response.BadRequest(w, "Invalid operator ID")
		return
	}

	operator, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Operator not found")
			return
		}
		response.InternalError(w, "Failed to get operator")
		return
	}

	tenantID, _ := middleware.GetTenantUUID(r.Context())
	if operator.TenantID != tenantID {
		response.NotFound(w, "Operator not found")
		return
	}

	response.OK(w, dto.NewOperatorResponse(operator))
}

// List handles GET /api/v1/operators
func (h *OperatorHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantUUID(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired, "X-Tenant-ID required")
		return
	}

	operators, err := h.service.ListByTenant(r.Context(), tenantID)
	if err != nil {
		response.InternalError(w, "Failed to list operators")
		return
	}

	items := make([]dto.OperatorResponse, len(operators))
	for i, op := range operators {
		items[i] = dto.NewOperatorResponse(op)
	}

	pagination := dto.ParsePagination(r)
	response.OK(w, dto.OperatorListResponse{
		Operators: items,
		Meta:      dto.NewListMeta(pagination.Page, pagination.PerPage, len(items)),
	})
}

// Update handles PUT /api/v1/operators/{id}
func (h *OperatorHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := dto.ParseUUIDParam(r, "id")
	if err != nil {
		response.BadRequest(w, "Invalid operator ID")
		return
	}

	req, err := dto.ParseJSON[dto.UpdateOperatorRequest](r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Verify tenant match
	operator, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Operator not found")
			return
		}
		response.InternalError(w, "Failed to get operator")
		return
	}

	tenantID, _ := middleware.GetTenantUUID(r.Context())
	if operator.TenantID != tenantID {
		response.NotFound(w, "Operator not found")
		return
	}

	updated, err := h.service.Update(r.Context(), id, domain.OperatorRole(req.Role))
	if err != nil {
		response.InternalError(w, "Failed to update operator")
		return
	}

	response.OK(w, dto.NewOperatorResponse(updated))
}

// Delete handles DELETE /api/v1/operators/{id}
func (h *OperatorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := dto.ParseUUIDParam(r, "id")
	if err != nil {
		response.BadRequest(w, "Invalid operator ID")
		return
	}

	operator, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Operator not found")
			return
		}
		response.InternalError(w, "Failed to get operator")
		return
	}

	tenantID, _ := middleware.GetTenantUUID(r.Context())
	if operator.TenantID != tenantID {
		response.NotFound(w, "Operator not found")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		response.InternalError(w, "Failed to delete operator")
		return
	}

	response.NoContent(w)
}
