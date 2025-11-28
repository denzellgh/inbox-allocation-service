package handler

import (
	"net/http"

	"github.com/inbox-allocation-service/internal/api/dto"
	"github.com/inbox-allocation-service/internal/api/middleware"
	"github.com/inbox-allocation-service/internal/api/response"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/service"
)

type TenantHandler struct {
	service *service.TenantService
}

func NewTenantHandler(svc *service.TenantService) *TenantHandler {
	return &TenantHandler{service: svc}
}

// Get handles GET /api/v1/tenant
func (h *TenantHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantUUID(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired, "X-Tenant-ID required")
		return
	}

	tenant, err := h.service.GetByID(r.Context(), tenantID)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Tenant not found")
			return
		}
		response.InternalError(w, "Failed to get tenant")
		return
	}

	response.OK(w, dto.NewTenantResponse(tenant))
}

// UpdateWeights handles PUT /api/v1/tenant/weights
func (h *TenantHandler) UpdateWeights(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantUUID(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired, "X-Tenant-ID required")
		return
	}

	req, err := dto.ParseJSON[dto.UpdateTenantWeightsRequest](r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	operatorID, _ := middleware.GetOperatorUUID(r.Context())
	alpha, beta := req.ToDecimal()

	tenant, err := h.service.UpdateWeights(r.Context(), tenantID, alpha, beta, &operatorID)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Tenant not found")
			return
		}
		response.InternalError(w, "Failed to update weights")
		return
	}

	response.OK(w, dto.NewTenantResponse(tenant))
}
