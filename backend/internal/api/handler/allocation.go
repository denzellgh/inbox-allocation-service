package handler

import (
	"errors"
	"net/http"

	"github.com/inbox-allocation-service/internal/api/dto"
	"github.com/inbox-allocation-service/internal/api/middleware"
	"github.com/inbox-allocation-service/internal/api/response"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/service"
)

type AllocationHandler struct {
	service *service.AllocationService
}

func NewAllocationHandler(svc *service.AllocationService) *AllocationHandler {
	return &AllocationHandler{service: svc}
}

// Allocate handles POST /api/v1/allocate
func (h *AllocationHandler) Allocate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := middleware.GetTenantUUID(ctx)
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired, "X-Tenant-ID required")
		return
	}

	operatorID, ok := middleware.GetOperatorUUID(ctx)
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeOperatorRequired, "X-Operator-ID required")
		return
	}

	// Parse request (no body needed)
	req := dto.ParseAllocateRequest(r)
	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Execute allocation
	conv, err := h.service.Allocate(ctx, tenantID, operatorID)
	if err != nil {
		h.handleAllocationError(w, err)
		return
	}

	// Build response
	resp := dto.NewAllocationResponse(conv)
	response.OK(w, resp)
}

// Claim handles POST /api/v1/claim
func (h *AllocationHandler) Claim(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := middleware.GetTenantUUID(ctx)
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired, "X-Tenant-ID required")
		return
	}

	operatorID, ok := middleware.GetOperatorUUID(ctx)
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeOperatorRequired, "X-Operator-ID required")
		return
	}

	// Parse request
	req, err := dto.ParseClaimRequest(r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Execute claim
	conv, err := h.service.Claim(ctx, tenantID, operatorID, req.ConversationID)
	if err != nil {
		h.handleClaimError(w, err)
		return
	}

	// Build response
	resp := dto.NewAllocationResponse(conv)
	response.OK(w, resp)
}

// ==================== Error Handling ====================

func (h *AllocationHandler) handleAllocationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrOperatorNotAvailable):
		response.Error(w, http.StatusBadRequest, dto.ErrCodeOperatorNotAvailable,
			"Operator must be AVAILABLE to allocate conversations")
	case errors.Is(err, service.ErrNoSubscriptions):
		response.Error(w, http.StatusBadRequest, dto.ErrCodeNoSubscriptions,
			"Operator has no inbox subscriptions")
	case errors.Is(err, service.ErrNoConversationsAvailable):
		response.Error(w, http.StatusNotFound, dto.ErrCodeNoConversationsAvailable,
			"No conversations available for allocation")
	default:
		response.InternalError(w, "Failed to allocate conversation")
	}
}

func (h *AllocationHandler) handleClaimError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrOperatorNotAvailable):
		response.Error(w, http.StatusBadRequest, dto.ErrCodeOperatorNotAvailable,
			"Operator must be AVAILABLE to claim conversations")
	case errors.Is(err, service.ErrConversationNotQueued):
		response.Error(w, http.StatusConflict, dto.ErrCodeConversationNotQueued,
			"Conversation is not available for claim")
	case errors.Is(err, service.ErrConversationAlreadyClaimed):
		response.Error(w, http.StatusConflict, dto.ErrCodeConversationAlreadyClaimed,
			"This conversation has already been claimed by another operator")
	case errors.Is(err, service.ErrNotSubscribedToInbox):
		response.Error(w, http.StatusForbidden, dto.ErrCodeNotSubscribedToInbox,
			"You are not subscribed to this conversation's inbox")
	case errors.Is(err, domain.ErrNotFound):
		response.NotFound(w, "Conversation not found")
	default:
		response.InternalError(w, "Failed to claim conversation")
	}
}
