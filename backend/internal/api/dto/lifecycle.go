package dto

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

// ==================== Resolve Request ====================

type ResolveRequest struct {
	ConversationID uuid.UUID `json:"conversation_id"`
}

func ParseResolveRequest(r *http.Request) (*ResolveRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var req ResolveRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *ResolveRequest) Validate() []string {
	var errs []string
	if r.ConversationID == uuid.Nil {
		errs = append(errs, "conversation_id is required")
	}
	return errs
}

// ==================== Deallocate Request ====================

type DeallocateRequest struct {
	ConversationID uuid.UUID `json:"conversation_id"`
}

func ParseDeallocateRequest(r *http.Request) (*DeallocateRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var req DeallocateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *DeallocateRequest) Validate() []string {
	var errs []string
	if r.ConversationID == uuid.Nil {
		errs = append(errs, "conversation_id is required")
	}
	return errs
}

// ==================== Reassign Request ====================

type ReassignRequest struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	OperatorID     uuid.UUID `json:"operator_id"`
}

func ParseReassignRequest(r *http.Request) (*ReassignRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var req ReassignRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *ReassignRequest) Validate() []string {
	var errs []string
	if r.ConversationID == uuid.Nil {
		errs = append(errs, "conversation_id is required")
	}
	if r.OperatorID == uuid.Nil {
		errs = append(errs, "operator_id is required")
	}
	return errs
}

// ==================== Move Inbox Request ====================

type MoveInboxRequest struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	InboxID        uuid.UUID `json:"inbox_id"`
}

func ParseMoveInboxRequest(r *http.Request) (*MoveInboxRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var req MoveInboxRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *MoveInboxRequest) Validate() []string {
	var errs []string
	if r.ConversationID == uuid.Nil {
		errs = append(errs, "conversation_id is required")
	}
	if r.InboxID == uuid.Nil {
		errs = append(errs, "inbox_id is required")
	}
	return errs
}

// ==================== Lifecycle Response ====================

type LifecycleResponse struct {
	ID                     uuid.UUID  `json:"id"`
	TenantID               uuid.UUID  `json:"tenant_id"`
	InboxID                uuid.UUID  `json:"inbox_id"`
	ExternalConversationID string     `json:"external_conversation_id"`
	CustomerPhoneNumber    string     `json:"customer_phone_number"`
	State                  string     `json:"state"`
	AssignedOperatorID     *uuid.UUID `json:"assigned_operator_id"`
	LastMessageAt          string     `json:"last_message_at"`
	MessageCount           int        `json:"message_count"`
	PriorityScore          float64    `json:"priority_score"`
	CreatedAt              string     `json:"created_at"`
	UpdatedAt              string     `json:"updated_at"`
	ResolvedAt             *string    `json:"resolved_at"`
}

func NewLifecycleResponse(c *domain.ConversationRef) LifecycleResponse {
	priorityScore, _ := c.PriorityScore.Float64()

	var resolvedAt *string
	if c.ResolvedAt != nil {
		t := c.ResolvedAt.Format("2006-01-02T15:04:05Z07:00")
		resolvedAt = &t
	}

	return LifecycleResponse{
		ID:                     c.ID,
		TenantID:               c.TenantID,
		InboxID:                c.InboxID,
		ExternalConversationID: c.ExternalConversationID,
		CustomerPhoneNumber:    c.CustomerPhoneNumber,
		State:                  string(c.State),
		AssignedOperatorID:     c.AssignedOperatorID,
		LastMessageAt:          c.LastMessageAt.Format("2006-01-02T15:04:05Z07:00"),
		MessageCount:           int(c.MessageCount),
		PriorityScore:          priorityScore,
		CreatedAt:              c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:              c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ResolvedAt:             resolvedAt,
	}
}

// ==================== Error Codes ====================

const (
	ErrCodeConversationNotFound           = "CONVERSATION_NOT_FOUND"
	ErrCodeConversationNotAllocated       = "CONVERSATION_NOT_ALLOCATED"
	ErrCodeConversationAlreadyResolved    = "CONVERSATION_ALREADY_RESOLVED"
	ErrCodeInsufficientPermissions        = "INSUFFICIENT_PERMISSIONS"
	ErrCodeOperatorNotFoundLifecycle      = "OPERATOR_NOT_FOUND"
	ErrCodeOperatorNotSubscribedLifecycle = "OPERATOR_NOT_SUBSCRIBED"
	ErrCodeInboxNotFound                  = "INBOX_NOT_FOUND"
	ErrCodeInboxDifferentTenant           = "INBOX_DIFFERENT_TENANT"
)
