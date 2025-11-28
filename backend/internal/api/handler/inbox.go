package handler

import (
	"net/http"

	"github.com/inbox-allocation-service/internal/api/dto"
	"github.com/inbox-allocation-service/internal/api/middleware"
	"github.com/inbox-allocation-service/internal/api/response"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/service"
)

type InboxHandler struct {
	service *service.InboxService
}

func NewInboxHandler(svc *service.InboxService) *InboxHandler {
	return &InboxHandler{service: svc}
}

// ListForOperator handles GET /api/v1/inboxes
func (h *InboxHandler) ListForOperator(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantUUID(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired, "X-Tenant-ID required")
		return
	}

	var inboxes []*domain.Inbox
	var err error

	operatorID, hasOperator := middleware.GetOperatorUUID(r.Context())
	if hasOperator {
		inboxes, err = h.service.ListForOperator(r.Context(), tenantID, operatorID)
	} else {
		inboxes, err = h.service.ListByTenant(r.Context(), tenantID)
	}

	if err != nil {
		response.InternalError(w, "Failed to list inboxes")
		return
	}

	items := make([]dto.InboxResponse, len(inboxes))
	for i, inbox := range inboxes {
		items[i] = dto.NewInboxResponse(inbox)
	}

	pagination := dto.ParsePagination(r)
	response.OK(w, dto.InboxListResponse{
		Inboxes: items,
		Meta:    dto.NewListMeta(pagination.Page, pagination.PerPage, len(items)),
	})
}

// Create handles POST /api/v1/inboxes
func (h *InboxHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantUUID(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, response.ErrCodeTenantRequired, "X-Tenant-ID required")
		return
	}

	req, err := dto.ParseJSON[dto.CreateInboxRequest](r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	inbox, err := h.service.Create(r.Context(), tenantID, req.PhoneNumber, req.DisplayName)
	if err != nil {
		if err == domain.ErrAlreadyExists {
			response.Conflict(w, response.ErrCodeConflict, "Phone number already exists")
			return
		}
		response.InternalError(w, "Failed to create inbox")
		return
	}

	response.Created(w, dto.NewInboxResponse(inbox))
}

// GetByID handles GET /api/v1/inboxes/{id}
func (h *InboxHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := dto.ParseUUIDParam(r, "id")
	if err != nil {
		response.BadRequest(w, "Invalid inbox ID")
		return
	}

	inbox, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Inbox not found")
			return
		}
		response.InternalError(w, "Failed to get inbox")
		return
	}

	tenantID, _ := middleware.GetTenantUUID(r.Context())
	if inbox.TenantID != tenantID {
		response.NotFound(w, "Inbox not found")
		return
	}

	response.OK(w, dto.NewInboxResponse(inbox))
}

// Update handles PUT /api/v1/inboxes/{id}
func (h *InboxHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := dto.ParseUUIDParam(r, "id")
	if err != nil {
		response.BadRequest(w, "Invalid inbox ID")
		return
	}

	req, err := dto.ParseJSON[dto.UpdateInboxRequest](r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	existing, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Inbox not found")
			return
		}
		response.InternalError(w, "Failed to get inbox")
		return
	}

	tenantID, _ := middleware.GetTenantUUID(r.Context())
	if existing.TenantID != tenantID {
		response.NotFound(w, "Inbox not found")
		return
	}

	inbox, err := h.service.Update(r.Context(), id, req.PhoneNumber, req.DisplayName)
	if err != nil {
		if err == domain.ErrAlreadyExists {
			response.Conflict(w, response.ErrCodeConflict, "Phone number already exists")
			return
		}
		response.InternalError(w, "Failed to update inbox")
		return
	}

	response.OK(w, dto.NewInboxResponse(inbox))
}

// Delete handles DELETE /api/v1/inboxes/{id}
func (h *InboxHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := dto.ParseUUIDParam(r, "id")
	if err != nil {
		response.BadRequest(w, "Invalid inbox ID")
		return
	}

	inbox, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Inbox not found")
			return
		}
		response.InternalError(w, "Failed to get inbox")
		return
	}

	tenantID, _ := middleware.GetTenantUUID(r.Context())
	if inbox.TenantID != tenantID {
		response.NotFound(w, "Inbox not found")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		response.InternalError(w, "Failed to delete inbox")
		return
	}

	response.NoContent(w)
}
