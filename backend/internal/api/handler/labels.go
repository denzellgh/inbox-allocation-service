package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/api/dto"
	"github.com/inbox-allocation-service/internal/api/middleware"
	"github.com/inbox-allocation-service/internal/api/response"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/service"
)

type LabelHandler struct {
	service *service.LabelService
}

func NewLabelHandler(svc *service.LabelService) *LabelHandler {
	return &LabelHandler{service: svc}
}

// Create handles POST /api/v1/labels
func (h *LabelHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	req, err := dto.ParseCreateLabelRequest(r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Execute
	label, err := h.service.CreateLabel(ctx, tenantID, operatorID, req.InboxID, role, req.Name, req.Color)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Created(w, dto.NewLabelResponse(label))
}

// List handles GET /api/v1/labels?inbox_id=
func (h *LabelHandler) List(w http.ResponseWriter, r *http.Request) {
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

	// Parse inbox_id from query
	inboxIDStr := r.URL.Query().Get("inbox_id")
	if inboxIDStr == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_QUERY", "inbox_id query parameter is required")
		return
	}

	inboxID, err := uuid.Parse(inboxIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_QUERY", "inbox_id must be a valid UUID")
		return
	}

	// Execute
	labels, err := h.service.ListLabelsByInbox(ctx, tenantID, operatorID, inboxID, role)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, dto.NewLabelListResponse(labels))
}

// Update handles PUT /api/v1/labels/{id}
func (h *LabelHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	// Parse label ID from path
	labelIDStr := chi.URLParam(r, "id")
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_PATH", "id must be a valid UUID")
		return
	}

	// Parse request
	req, err := dto.ParseUpdateLabelRequest(r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Execute
	label, err := h.service.UpdateLabel(ctx, tenantID, operatorID, labelID, role, req.Name, req.Color)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, dto.NewLabelResponse(label))
}

// Delete handles DELETE /api/v1/labels/{id}
func (h *LabelHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	// Parse label ID from path
	labelIDStr := chi.URLParam(r, "id")
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_PATH", "id must be a valid UUID")
		return
	}

	// Execute
	if err := h.service.DeleteLabel(ctx, tenantID, operatorID, labelID, role); err != nil {
		h.handleError(w, err)
		return
	}

	response.NoContent(w)
}

// Attach handles POST /api/v1/labels/attach
func (h *LabelHandler) Attach(w http.ResponseWriter, r *http.Request) {
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
	req, err := dto.ParseAttachLabelRequest(r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Execute
	if err := h.service.AttachLabelToConversation(ctx, tenantID, operatorID, req.ConversationID, req.LabelID, role); err != nil {
		h.handleError(w, err)
		return
	}

	response.NoContent(w)
}

// Detach handles POST /api/v1/labels/detach
func (h *LabelHandler) Detach(w http.ResponseWriter, r *http.Request) {
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
	req, err := dto.ParseDetachLabelRequest(r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	// Execute
	if err := h.service.DetachLabelFromConversation(ctx, tenantID, operatorID, req.ConversationID, req.LabelID, role); err != nil {
		h.handleError(w, err)
		return
	}

	response.NoContent(w)
}

// ==================== Error Handling ====================

func (h *LabelHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrLabelNotFound):
		response.Error(w, http.StatusNotFound, dto.ErrCodeLabelNotFound,
			"Label not found")
	case errors.Is(err, domain.ErrNotFound):
		response.Error(w, http.StatusNotFound, dto.ErrCodeConversationNotFound,
			"Resource not found")
	case errors.Is(err, service.ErrLabelNameConflict):
		response.Error(w, http.StatusConflict, dto.ErrCodeLabelNameConflict,
			"A label with this name already exists in this inbox")
	case errors.Is(err, service.ErrLabelInboxMismatch):
		response.Error(w, http.StatusBadRequest, dto.ErrCodeLabelInboxMismatch,
			"Label inbox does not match conversation inbox")
	case errors.Is(err, service.ErrLabelPermissionDenied):
		response.Error(w, http.StatusForbidden, dto.ErrCodeLabelPermissionDenied,
			"You don't have permission for this operation")
	default:
		response.InternalError(w, "Failed to process label operation")
	}
}
