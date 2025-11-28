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

type LifecycleHandler struct {
	service *service.LifecycleService
}

func NewLifecycleHandler(svc *service.LifecycleService) *LifecycleHandler {
	return &LifecycleHandler{service: svc}
}

// Resolve handles POST /api/v1/resolve
func (h *LifecycleHandler) Resolve(w http.ResponseWriter, r *http.Request) {
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

	role, _ := middleware.GetOperatorRole(ctx)

	// Parse request
	req, err := dto.ParseResolveRequest(r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Execute
	conv, err := h.service.Resolve(ctx, tenantID, operatorID, req.ConversationID, role)
	if err != nil {
		h.handleError(w, err, "resolve")
		return
	}

	response.OK(w, dto.NewLifecycleResponse(conv))
}

// Deallocate handles POST /api/v1/deallocate
func (h *LifecycleHandler) Deallocate(w http.ResponseWriter, r *http.Request) {
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

	role, _ := middleware.GetOperatorRole(ctx)

	// Parse request
	req, err := dto.ParseDeallocateRequest(r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Execute
	conv, err := h.service.Deallocate(ctx, tenantID, operatorID, req.ConversationID, role)
	if err != nil {
		h.handleError(w, err, "deallocate")
		return
	}

	response.OK(w, dto.NewLifecycleResponse(conv))
}

// Reassign handles POST /api/v1/reassign
func (h *LifecycleHandler) Reassign(w http.ResponseWriter, r *http.Request) {
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

	role, _ := middleware.GetOperatorRole(ctx)

	// Parse request
	req, err := dto.ParseReassignRequest(r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Execute
	conv, err := h.service.Reassign(ctx, tenantID, operatorID, req.ConversationID, req.OperatorID, role)
	if err != nil {
		h.handleError(w, err, "reassign")
		return
	}

	response.OK(w, dto.NewLifecycleResponse(conv))
}

// MoveInbox handles POST /api/v1/move_inbox
func (h *LifecycleHandler) MoveInbox(w http.ResponseWriter, r *http.Request) {
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

	role, _ := middleware.GetOperatorRole(ctx)

	// Parse request
	req, err := dto.ParseMoveInboxRequest(r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Execute
	conv, err := h.service.MoveInbox(ctx, tenantID, operatorID, req.ConversationID, req.InboxID, role)
	if err != nil {
		h.handleError(w, err, "move_inbox")
		return
	}

	response.OK(w, dto.NewLifecycleResponse(conv))
}

// ==================== Error Handling ====================

func (h *LifecycleHandler) handleError(w http.ResponseWriter, err error, operation string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.Error(w, http.StatusNotFound, dto.ErrCodeConversationNotFound,
			"Conversation not found")
	case errors.Is(err, service.ErrConversationNotAllocated):
		response.Error(w, http.StatusConflict, dto.ErrCodeConversationNotAllocated,
			"Conversation is not in ALLOCATED state")
	case errors.Is(err, service.ErrConversationAlreadyResolved):
		response.Error(w, http.StatusConflict, dto.ErrCodeConversationAlreadyResolved,
			"Conversation is already resolved")
	case errors.Is(err, service.ErrInsufficientPermissions):
		response.Error(w, http.StatusForbidden, dto.ErrCodeInsufficientPermissions,
			"You don't have permission for this operation")
	case errors.Is(err, service.ErrTargetOperatorNotFound):
		response.Error(w, http.StatusNotFound, dto.ErrCodeOperatorNotFoundLifecycle,
			"Target operator not found")
	case errors.Is(err, service.ErrTargetOperatorNotSubscribed):
		response.Error(w, http.StatusBadRequest, dto.ErrCodeOperatorNotSubscribedLifecycle,
			"Target operator is not subscribed to the inbox")
	case errors.Is(err, service.ErrTargetInboxNotFound):
		response.Error(w, http.StatusNotFound, dto.ErrCodeInboxNotFound,
			"Target inbox not found")
	case errors.Is(err, service.ErrTargetInboxDifferentTenant):
		response.Error(w, http.StatusBadRequest, dto.ErrCodeInboxDifferentTenant,
			"Target inbox belongs to a different tenant")
	default:
		response.InternalError(w, "Failed to "+operation+" conversation")
	}
}
