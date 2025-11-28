package dto

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

// ==================== Constants ====================

const (
	SortNewest   = "newest"
	SortOldest   = "oldest"
	SortPriority = "priority"

	MaxConversationsPerQuery = 100
	DefaultPerPage           = 50
)

// ==================== Cursor ====================

type Cursor struct {
	Timestamp time.Time `json:"ts"`
	ID        uuid.UUID `json:"id"`
}

func EncodeCursor(ts time.Time, id uuid.UUID) string {
	c := Cursor{Timestamp: ts, ID: id}
	data, _ := json.Marshal(c)
	return base64.URLEncoding.EncodeToString(data)
}

func DecodeCursor(encoded string) (*Cursor, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	var c Cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// ==================== List Request ====================

type ListConversationsRequest struct {
	// Filters
	State      *string    `json:"state,omitempty"`
	InboxID    *uuid.UUID `json:"inbox_id,omitempty"`
	OperatorID *uuid.UUID `json:"operator_id,omitempty"`
	LabelID    *uuid.UUID `json:"label_id,omitempty"`

	// Sorting
	Sort string `json:"sort"`

	// Pagination
	Cursor  string `json:"cursor,omitempty"`
	PerPage int    `json:"per_page"`
}

func ParseListConversationsRequest(r *http.Request) *ListConversationsRequest {
	req := &ListConversationsRequest{
		Sort:    r.URL.Query().Get("sort"),
		Cursor:  r.URL.Query().Get("cursor"),
		PerPage: DefaultPerPage,
	}

	// Parse state filter
	if state := r.URL.Query().Get("state"); state != "" {
		req.State = &state
	}

	// Parse inbox_id filter
	if inboxIDStr := r.URL.Query().Get("inbox_id"); inboxIDStr != "" {
		if id, err := uuid.Parse(inboxIDStr); err == nil {
			req.InboxID = &id
		}
	}

	// Parse operator_id filter
	if opIDStr := r.URL.Query().Get("operator_id"); opIDStr != "" {
		if id, err := uuid.Parse(opIDStr); err == nil {
			req.OperatorID = &id
		}
	}

	// Parse label_id filter
	if labelIDStr := r.URL.Query().Get("label_id"); labelIDStr != "" {
		if id, err := uuid.Parse(labelIDStr); err == nil {
			req.LabelID = &id
		}
	}

	// Normalize sort
	if req.Sort == "" {
		req.Sort = SortNewest
	}

	// Parse per_page
	pagination := ParsePagination(r)
	req.PerPage = pagination.PerPage
	if req.PerPage > MaxConversationsPerQuery {
		req.PerPage = MaxConversationsPerQuery
	}

	return req
}

func (r *ListConversationsRequest) Validate() []string {
	var errs []string

	// Validate state
	if r.State != nil {
		state := domain.ConversationState(*r.State)
		if !state.IsValid() {
			errs = append(errs, "state must be QUEUED, ALLOCATED, or RESOLVED")
		}
	}

	// Validate sort
	sort := strings.ToLower(r.Sort)
	if sort != SortNewest && sort != SortOldest && sort != SortPriority {
		errs = append(errs, "sort must be newest, oldest, or priority")
	}

	return errs
}

func (r *ListConversationsRequest) GetCursor() *Cursor {
	if r.Cursor == "" {
		return nil
	}
	cursor, err := DecodeCursor(r.Cursor)
	if err != nil {
		return nil
	}
	return cursor
}

// ==================== Search Request ====================

type SearchConversationsRequest struct {
	Phone string `json:"phone"`
}

func ParseSearchRequest(r *http.Request) *SearchConversationsRequest {
	return &SearchConversationsRequest{
		Phone: r.URL.Query().Get("phone"),
	}
}

func (r *SearchConversationsRequest) Validate() []string {
	var errs []string
	if strings.TrimSpace(r.Phone) == "" {
		errs = append(errs, "phone is required")
	}
	return errs
}

// Normalize phone for search (remove spaces, ensure + prefix for international)
func (r *SearchConversationsRequest) NormalizedPhone() string {
	phone := strings.TrimSpace(r.Phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	return phone
}

// ==================== Conversation Response ====================

type ConversationResponse struct {
	ID                     uuid.UUID      `json:"id"`
	TenantID               uuid.UUID      `json:"tenant_id"`
	InboxID                uuid.UUID      `json:"inbox_id"`
	ExternalConversationID string         `json:"external_conversation_id"`
	CustomerPhoneNumber    string         `json:"customer_phone_number"`
	State                  string         `json:"state"`
	AssignedOperatorID     *uuid.UUID     `json:"assigned_operator_id"`
	LastMessageAt          time.Time      `json:"last_message_at"`
	MessageCount           int            `json:"message_count"`
	PriorityScore          float64        `json:"priority_score"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	ResolvedAt             *time.Time     `json:"resolved_at"`
	Labels                 []LabelSummary `json:"labels,omitempty"`
}

type LabelSummary struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color *string   `json:"color,omitempty"`
}

func NewConversationResponse(c *domain.ConversationRef) ConversationResponse {
	priorityScore, _ := c.PriorityScore.Float64()
	return ConversationResponse{
		ID:                     c.ID,
		TenantID:               c.TenantID,
		InboxID:                c.InboxID,
		ExternalConversationID: c.ExternalConversationID,
		CustomerPhoneNumber:    c.CustomerPhoneNumber,
		State:                  string(c.State),
		AssignedOperatorID:     c.AssignedOperatorID,
		LastMessageAt:          c.LastMessageAt,
		MessageCount:           int(c.MessageCount),
		PriorityScore:          priorityScore,
		CreatedAt:              c.CreatedAt,
		UpdatedAt:              c.UpdatedAt,
		ResolvedAt:             c.ResolvedAt,
		Labels:                 []LabelSummary{}, // Populated separately if needed
	}
}

func NewConversationResponseWithLabels(c *domain.ConversationRef, labels []*domain.Label) ConversationResponse {
	resp := NewConversationResponse(c)
	resp.Labels = make([]LabelSummary, len(labels))
	for i, l := range labels {
		resp.Labels[i] = LabelSummary{
			ID:    l.ID,
			Name:  l.Name,
			Color: l.Color,
		}
	}
	return resp
}

// ==================== List Response ====================

type ConversationListMeta struct {
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
	Count      int    `json:"count"`
}

type ConversationListResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Meta          ConversationListMeta   `json:"meta"`
}

func NewConversationListResponse(conversations []*domain.ConversationRef, perPage int) ConversationListResponse {
	items := make([]ConversationResponse, len(conversations))
	for i, c := range conversations {
		items[i] = NewConversationResponse(c)
	}

	resp := ConversationListResponse{
		Conversations: items,
		Meta: ConversationListMeta{
			Count:   len(items),
			HasMore: len(items) >= perPage,
		},
	}

	// Generate next cursor from last item
	if len(conversations) > 0 && resp.Meta.HasMore {
		last := conversations[len(conversations)-1]
		resp.Meta.NextCursor = EncodeCursor(last.LastMessageAt, last.ID)
	}

	return resp
}

// ==================== Search Response ====================

type SearchMeta struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}

type SearchConversationsResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Meta          SearchMeta             `json:"meta"`
}

func NewSearchResponse(conversations []*domain.ConversationRef, query string) SearchConversationsResponse {
	items := make([]ConversationResponse, len(conversations))
	for i, c := range conversations {
		items[i] = NewConversationResponse(c)
	}
	return SearchConversationsResponse{
		Conversations: items,
		Meta: SearchMeta{
			Query: query,
			Count: len(items),
		},
	}
}
