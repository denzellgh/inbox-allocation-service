package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

type CreateInboxRequest struct {
	PhoneNumber string `json:"phone_number"`
	DisplayName string `json:"display_name"`
}

func (r *CreateInboxRequest) Validate() []string {
	var errs []string
	if err := ValidateRequired(r.PhoneNumber, "phone_number"); err != nil {
		errs = append(errs, err.Error())
	}
	if err := ValidateMaxLength(r.PhoneNumber, 20, "phone_number"); err != nil {
		errs = append(errs, err.Error())
	}
	if err := ValidateRequired(r.DisplayName, "display_name"); err != nil {
		errs = append(errs, err.Error())
	}
	if err := ValidateMaxLength(r.DisplayName, 255, "display_name"); err != nil {
		errs = append(errs, err.Error())
	}
	return errs
}

type UpdateInboxRequest struct {
	PhoneNumber *string `json:"phone_number,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
}

func (r *UpdateInboxRequest) Validate() []string {
	var errs []string
	if r.PhoneNumber != nil {
		if err := ValidateMaxLength(*r.PhoneNumber, 20, "phone_number"); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if r.DisplayName != nil {
		if err := ValidateMaxLength(*r.DisplayName, 255, "display_name"); err != nil {
			errs = append(errs, err.Error())
		}
	}
	return errs
}

type InboxResponse struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	PhoneNumber string    `json:"phone_number"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewInboxResponse(inbox *domain.Inbox) InboxResponse {
	return InboxResponse{
		ID:          inbox.ID,
		TenantID:    inbox.TenantID,
		PhoneNumber: inbox.PhoneNumber,
		DisplayName: inbox.DisplayName,
		CreatedAt:   inbox.CreatedAt,
		UpdatedAt:   inbox.UpdatedAt,
	}
}

type InboxListResponse struct {
	Inboxes []InboxResponse `json:"inboxes"`
	Meta    ListMeta        `json:"meta"`
}
