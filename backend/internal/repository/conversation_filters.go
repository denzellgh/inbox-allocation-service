package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

// ConversationFilters holds all filter options for listing conversations
type ConversationFilters struct {
	// Required
	TenantID uuid.UUID

	// Optional filters
	State      *domain.ConversationState
	InboxID    *uuid.UUID
	OperatorID *uuid.UUID
	LabelID    *uuid.UUID

	// Access control - if set, only return conversations in these inboxes
	AllowedInboxIDs []uuid.UUID

	// Sorting: "newest", "oldest", "priority"
	SortOrder string

	// Cursor pagination
	CursorTimestamp *time.Time
	CursorID        *uuid.UUID

	// Limit
	Limit int
}

// HasCursor returns true if cursor pagination is active
func (f *ConversationFilters) HasCursor() bool {
	return f.CursorTimestamp != nil && f.CursorID != nil
}

// GetLimit returns the limit, defaulting to 50
func (f *ConversationFilters) GetLimit() int {
	if f.Limit <= 0 {
		return 50
	}
	if f.Limit > 100 {
		return 100
	}
	return f.Limit
}
