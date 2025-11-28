package handler

import (
	"net/http"

	"github.com/inbox-allocation-service/internal/api/dto"
	"github.com/inbox-allocation-service/internal/api/middleware"
	"github.com/inbox-allocation-service/internal/api/response"
	"github.com/inbox-allocation-service/internal/service"
)

type SubscriptionHandler struct {
	subSvc      *service.SubscriptionService
	operatorSvc *service.OperatorService
	inboxSvc    *service.InboxService
}

func NewSubscriptionHandler(subSvc *service.SubscriptionService, opSvc *service.OperatorService, inSvc *service.InboxService) *SubscriptionHandler {
	return &SubscriptionHandler{subSvc: subSvc, operatorSvc: opSvc, inboxSvc: inSvc}
}

// Subscribe handles POST /api/v1/inboxes/{inbox_id}/operators
func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	inboxID, err := dto.ParseUUIDParam(r, "inbox_id")
	if err != nil {
		response.BadRequest(w, "Invalid inbox ID")
		return
	}

	req, err := dto.ParseJSON[dto.SubscribeOperatorRequest](r)
	if err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, "Validation failed", errs...)
		return
	}

	tenantID, _ := middleware.GetTenantUUID(r.Context())

	// Verify inbox belongs to tenant
	inbox, err := h.inboxSvc.GetByID(r.Context(), inboxID)
	if err != nil || inbox.TenantID != tenantID {
		response.NotFound(w, "Inbox not found")
		return
	}

	// Verify operator belongs to tenant
	operator, err := h.operatorSvc.GetByID(r.Context(), req.OperatorID)
	if err != nil || operator.TenantID != tenantID {
		response.NotFound(w, "Operator not found")
		return
	}

	sub, err := h.subSvc.Subscribe(r.Context(), req.OperatorID, inboxID)
	if err != nil {
		response.InternalError(w, "Failed to subscribe")
		return
	}

	response.Created(w, dto.NewSubscriptionResponse(sub))
}

// Unsubscribe handles DELETE /api/v1/inboxes/{inbox_id}/operators/{operator_id}
func (h *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	inboxID, err := dto.ParseUUIDParam(r, "inbox_id")
	if err != nil {
		response.BadRequest(w, "Invalid inbox ID")
		return
	}

	operatorID, err := dto.ParseUUIDParam(r, "operator_id")
	if err != nil {
		response.BadRequest(w, "Invalid operator ID")
		return
	}

	if err := h.subSvc.Unsubscribe(r.Context(), operatorID, inboxID); err != nil {
		response.InternalError(w, "Failed to unsubscribe")
		return
	}

	response.NoContent(w)
}

// ListOperators handles GET /api/v1/inboxes/{inbox_id}/operators
func (h *SubscriptionHandler) ListOperators(w http.ResponseWriter, r *http.Request) {
	inboxID, err := dto.ParseUUIDParam(r, "inbox_id")
	if err != nil {
		response.BadRequest(w, "Invalid inbox ID")
		return
	}

	subs, err := h.subSvc.GetOperatorsByInbox(r.Context(), inboxID)
	if err != nil {
		response.InternalError(w, "Failed to list operators")
		return
	}

	items := make([]dto.SubscriptionResponse, len(subs))
	for i, sub := range subs {
		items[i] = dto.NewSubscriptionResponse(sub)
	}

	pagination := dto.ParsePagination(r)
	response.OK(w, dto.SubscriptionListResponse{
		Subscriptions: items,
		Meta:          dto.NewListMeta(pagination.Page, pagination.PerPage, len(items)),
	})
}

// ListInboxes handles GET /api/v1/operators/{operator_id}/inboxes
func (h *SubscriptionHandler) ListInboxes(w http.ResponseWriter, r *http.Request) {
	operatorID, err := dto.ParseUUIDParam(r, "operator_id")
	if err != nil {
		response.BadRequest(w, "Invalid operator ID")
		return
	}

	subs, err := h.subSvc.GetInboxesByOperator(r.Context(), operatorID)
	if err != nil {
		response.InternalError(w, "Failed to list inboxes")
		return
	}

	items := make([]dto.SubscriptionResponse, len(subs))
	for i, sub := range subs {
		items[i] = dto.NewSubscriptionResponse(sub)
	}

	pagination := dto.ParsePagination(r)
	response.OK(w, dto.SubscriptionListResponse{
		Subscriptions: items,
		Meta:          dto.NewListMeta(pagination.Page, pagination.PerPage, len(items)),
	})
}
