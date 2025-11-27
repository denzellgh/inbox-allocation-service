package dto

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

// ==================== Create Label Request ====================

type CreateLabelRequest struct {
	InboxID uuid.UUID `json:"inbox_id"`
	Name    string    `json:"name"`
	Color   *string   `json:"color"`
}

func ParseCreateLabelRequest(r *http.Request) (*CreateLabelRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var req CreateLabelRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *CreateLabelRequest) Validate() []string {
	var errs []string
	if r.InboxID == uuid.Nil {
		errs = append(errs, "inbox_id is required")
	}
	name := strings.TrimSpace(r.Name)
	if name == "" {
		errs = append(errs, "name is required")
	} else if len(name) > 64 {
		errs = append(errs, "name must be 64 characters or less")
	}
	if r.Color != nil && len(*r.Color) > 32 {
		errs = append(errs, "color must be 32 characters or less")
	}
	return errs
}

// ==================== Update Label Request ====================

type UpdateLabelRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

func ParseUpdateLabelRequest(r *http.Request) (*UpdateLabelRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var req UpdateLabelRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *UpdateLabelRequest) Validate() []string {
	var errs []string
	if r.Name == nil && r.Color == nil {
		errs = append(errs, "at least one field (name or color) must be provided")
		return errs
	}
	if r.Name != nil {
		name := strings.TrimSpace(*r.Name)
		if name == "" {
			errs = append(errs, "name cannot be empty")
		} else if len(name) > 64 {
			errs = append(errs, "name must be 64 characters or less")
		}
	}
	if r.Color != nil && len(*r.Color) > 32 {
		errs = append(errs, "color must be 32 characters or less")
	}
	return errs
}

// ==================== Attach Label Request ====================

type AttachLabelRequest struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	LabelID        uuid.UUID `json:"label_id"`
}

func ParseAttachLabelRequest(r *http.Request) (*AttachLabelRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var req AttachLabelRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *AttachLabelRequest) Validate() []string {
	var errs []string
	if r.ConversationID == uuid.Nil {
		errs = append(errs, "conversation_id is required")
	}
	if r.LabelID == uuid.Nil {
		errs = append(errs, "label_id is required")
	}
	return errs
}

// ==================== Detach Label Request ====================

type DetachLabelRequest struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	LabelID        uuid.UUID `json:"label_id"`
}

func ParseDetachLabelRequest(r *http.Request) (*DetachLabelRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var req DetachLabelRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *DetachLabelRequest) Validate() []string {
	var errs []string
	if r.ConversationID == uuid.Nil {
		errs = append(errs, "conversation_id is required")
	}
	if r.LabelID == uuid.Nil {
		errs = append(errs, "label_id is required")
	}
	return errs
}

// ==================== Label Response ====================

type LabelResponse struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	InboxID   uuid.UUID  `json:"inbox_id"`
	Name      string     `json:"name"`
	Color     *string    `json:"color"`
	CreatedBy *uuid.UUID `json:"created_by"`
	CreatedAt string     `json:"created_at"`
}

func NewLabelResponse(l *domain.Label) LabelResponse {
	return LabelResponse{
		ID:        l.ID,
		TenantID:  l.TenantID,
		InboxID:   l.InboxID,
		Name:      l.Name,
		Color:     l.Color,
		CreatedBy: l.CreatedBy,
		CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func NewLabelListResponse(labels []*domain.Label) []LabelResponse {
	result := make([]LabelResponse, len(labels))
	for i, l := range labels {
		result[i] = NewLabelResponse(l)
	}
	return result
}

// ==================== Error Codes ====================

const (
	ErrCodeLabelNotFound         = "LABEL_NOT_FOUND"
	ErrCodeLabelNameConflict     = "LABEL_NAME_CONFLICT"
	ErrCodeLabelInboxMismatch    = "LABEL_INBOX_MISMATCH"
	ErrCodeLabelPermissionDenied = "LABEL_PERMISSION_DENIED"
	ErrCodeInboxNotFoundLabel    = "INBOX_NOT_FOUND"
)
