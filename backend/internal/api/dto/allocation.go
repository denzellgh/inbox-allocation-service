package dto

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

// ==================== Allocate Request ====================

// AllocateRequest is intentionally empty - allocation is automatic
// Operator ID and Tenant ID come from headers/context
type AllocateRequest struct {
	// No body needed - allocation is based on operator context
}

func ParseAllocateRequest(r *http.Request) *AllocateRequest {
	return &AllocateRequest{}
}

func (r *AllocateRequest) Validate() []string {
	// No validation needed - operator context is validated by middleware
	return nil
}

// ==================== Claim Request ====================

type ClaimRequest struct {
	ConversationID uuid.UUID `json:"conversation_id"`
}

func ParseClaimRequest(r *http.Request) (*ClaimRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var req ClaimRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *ClaimRequest) Validate() []string {
	var errs []string

	if r.ConversationID == uuid.Nil {
		errs = append(errs, "conversation_id is required")
	}

	return errs
}

// ==================== Allocation Response ====================

type AllocationResponse struct {
	ID                     uuid.UUID  `json:"id"`
	TenantID               uuid.UUID  `json:"tenant_id"`
	InboxID                uuid.UUID  `json:"inbox_id"`
	ExternalConversationID string     `json:"external_conversation_id"`
	CustomerPhoneNumber    string     `json:"customer_phone_number"`
	State                  string     `json:"state"`
	AssignedOperatorID     uuid.UUID  `json:"assigned_operator_id"`
	LastMessageAt          string     `json:"last_message_at"`
	MessageCount           int        `json:"message_count"`
	PriorityScore          float64    `json:"priority_score"`
	CreatedAt              string     `json:"created_at"`
	UpdatedAt              string     `json:"updated_at"`
	ResolvedAt             *string    `json:"resolved_at"`
	AllocatedAt            string     `json:"allocated_at"`
}

func NewAllocationResponse(c *domain.ConversationRef) AllocationResponse {
	priorityScore, _ := c.PriorityScore.Float64()
	
	var resolvedAt *string
	if c.ResolvedAt != nil {
		t := c.ResolvedAt.Format("2006-01-02T15:04:05Z07:00")
		resolvedAt = &t
	}

	var assignedOperatorID uuid.UUID
	if c.AssignedOperatorID != nil {
		assignedOperatorID = *c.AssignedOperatorID
	}

	return AllocationResponse{
		ID:                     c.ID,
		TenantID:               c.TenantID,
		InboxID:                c.InboxID,
		ExternalConversationID: c.ExternalConversationID,
		CustomerPhoneNumber:    c.CustomerPhoneNumber,
		State:                  string(c.State),
		AssignedOperatorID:     assignedOperatorID,
		LastMessageAt:          c.LastMessageAt.Format("2006-01-02T15:04:05Z07:00"),
		MessageCount:           int(c.MessageCount),
		PriorityScore:          priorityScore,
		CreatedAt:              c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:              c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ResolvedAt:             resolvedAt,
		AllocatedAt:            c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), // Same as UpdatedAt for allocation time
	}
}

// ==================== Error Codes ====================

const (
	ErrCodeOperatorNotAvailable       = "OPERATOR_NOT_AVAILABLE"
	ErrCodeNoSubscriptions            = "NO_SUBSCRIPTIONS"
	ErrCodeNoConversationsAvailable   = "NO_CONVERSATIONS_AVAILABLE"
	ErrCodeConversationNotQueued      = "CONVERSATION_NOT_QUEUED"
	ErrCodeConversationAlreadyClaimed = "CONVERSATION_ALREADY_CLAIMED"
	ErrCodeNotSubscribedToInbox       = "NOT_SUBSCRIBED_TO_INBOX"
)
