package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

type SubscribeOperatorRequest struct {
	OperatorID uuid.UUID `json:"operator_id"`
}

func (r *SubscribeOperatorRequest) Validate() []string {
	var errs []string
	if r.OperatorID == uuid.Nil {
		errs = append(errs, "operator_id is required")
	}
	return errs
}

type SubscriptionResponse struct {
	ID         uuid.UUID `json:"id"`
	OperatorID uuid.UUID `json:"operator_id"`
	InboxID    uuid.UUID `json:"inbox_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewSubscriptionResponse(sub *domain.OperatorInboxSubscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:         sub.ID,
		OperatorID: sub.OperatorID,
		InboxID:    sub.InboxID,
		CreatedAt:  sub.CreatedAt,
	}
}

type SubscriptionListResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
	Meta          ListMeta               `json:"meta"`
}

type OperatorWithSubscription struct {
	ID           uuid.UUID `json:"id"`
	Role         string    `json:"role"`
	Status       string    `json:"status,omitempty"`
	SubscribedAt time.Time `json:"subscribed_at"`
}

type InboxOperatorsResponse struct {
	InboxID   uuid.UUID                  `json:"inbox_id"`
	Operators []OperatorWithSubscription `json:"operators"`
	Meta      ListMeta                   `json:"meta"`
}

type InboxWithSubscription struct {
	ID           uuid.UUID `json:"id"`
	DisplayName  string    `json:"display_name"`
	PhoneNumber  string    `json:"phone_number"`
	SubscribedAt time.Time `json:"subscribed_at"`
}

type OperatorInboxesResponse struct {
	OperatorID uuid.UUID               `json:"operator_id"`
	Inboxes    []InboxWithSubscription `json:"inboxes"`
	Meta       ListMeta                `json:"meta"`
}
