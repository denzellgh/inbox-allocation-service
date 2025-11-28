package handler

import (
	"net/http"

	"github.com/inbox-allocation-service/internal/api/dto"
	"github.com/inbox-allocation-service/internal/api/middleware"
	"github.com/inbox-allocation-service/internal/api/response"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/service"
)

type ConversationHandler struct {
	service *service.ConversationService
}

func NewConversationHandler(svc *service.ConversationService) *ConversationHandler {
	return &ConversationHandler{service: svc}
}

// List handles GET /api/v1/conversations
func (h *ConversationHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := middleware.GetTenantUUID(ctx)
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired, "X-Tenant-ID required")
		return
	}

	operatorID, _ := middleware.GetOperatorUUID(ctx)
	role, _ := middleware.GetOperatorRole(ctx)

	// Parse request
	req := dto.ParseListConversationsRequest(r)

	// Validate
	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Build params
	params := service.ListConversationsParams{
		TenantID:   tenantID,
		OperatorID: operatorID,
		Role:       role,
		Sort:       req.Sort,
		Cursor:     req.GetCursor(),
		PerPage:    req.PerPage,
	}

	// Apply filters
	if req.State != nil {
		state := domain.ConversationState(*req.State)
		params.State = &state
	}
	if req.InboxID != nil {
		params.InboxID = req.InboxID
	}
	if req.OperatorID != nil {
		params.OperatorFilterID = req.OperatorID
	}
	if req.LabelID != nil {
		params.LabelID = req.LabelID
	}

	// Execute
	conversations, err := h.service.List(ctx, params)
	if err != nil {
		response.InternalError(w, "Failed to list conversations")
		return
	}

	// Build response
	resp := dto.NewConversationListResponse(conversations, req.PerPage)
	response.OK(w, resp)
}

// GetByID handles GET /api/v1/conversations/{id}
func (h *ConversationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := middleware.GetTenantUUID(ctx)
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired, "X-Tenant-ID required")
		return
	}

	operatorID, _ := middleware.GetOperatorUUID(ctx)
	role, _ := middleware.GetOperatorRole(ctx)

	// Parse conversation ID
	conversationID, err := dto.ParseUUIDParam(r, "id")
	if err != nil {
		response.BadRequest(w, "Invalid conversation ID")
		return
	}

	// Get conversation
	conv, err := h.service.GetByID(ctx, tenantID, conversationID)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Conversation not found")
			return
		}
		response.InternalError(w, "Failed to get conversation")
		return
	}

	// Check access
	if !h.service.CanAccess(ctx, operatorID, role, conv) {
		response.NotFound(w, "Conversation not found")
		return
	}

	// Get labels
	labels, _ := h.service.GetLabels(ctx, conversationID)

	// Build response
	resp := dto.NewConversationResponseWithLabels(conv, labels)
	response.OK(w, resp)
}

// Search handles GET /api/v1/search
func (h *ConversationHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := middleware.GetTenantUUID(ctx)
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired, "X-Tenant-ID required")
		return
	}

	operatorID, _ := middleware.GetOperatorUUID(ctx)
	role, _ := middleware.GetOperatorRole(ctx)

	// Parse request
	req := dto.ParseSearchRequest(r)

	// Validate
	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Execute search
	phone := req.NormalizedPhone()
	conversations, err := h.service.SearchByPhone(ctx, tenantID, phone, operatorID, role)
	if err != nil {
		response.InternalError(w, "Failed to search conversations")
		return
	}

	// Build response
	resp := dto.NewSearchResponse(conversations, phone)
	response.OK(w, resp)
}
