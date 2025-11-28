package dto_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/api/dto"
)

func TestEncodeCursor(t *testing.T) {
	ts := time.Date(2025, 11, 27, 0, 0, 0, 0, time.UTC)
	id := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")

	encoded := dto.EncodeCursor(ts, id)
	if encoded == "" {
		t.Error("expected non-empty cursor")
	}

	decoded, err := dto.DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if !decoded.Timestamp.Equal(ts) {
		t.Errorf("timestamp mismatch: got %v, want %v", decoded.Timestamp, ts)
	}
	if decoded.ID != id {
		t.Errorf("id mismatch: got %v, want %v", decoded.ID, id)
	}
}

func TestDecodeCursor_Invalid(t *testing.T) {
	_, err := dto.DecodeCursor("invalid-cursor")
	if err == nil {
		t.Error("expected error for invalid cursor")
	}
}

func TestListConversationsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		state   *string
		sort    string
		wantErr bool
	}{
		{"valid newest", nil, "newest", false},
		{"valid oldest", nil, "oldest", false},
		{"valid priority", nil, "priority", false},
		{"valid QUEUED state", strPtr("QUEUED"), "newest", false},
		{"valid ALLOCATED state", strPtr("ALLOCATED"), "newest", false},
		{"valid RESOLVED state", strPtr("RESOLVED"), "newest", false},
		{"invalid state", strPtr("INVALID"), "newest", true},
		{"invalid sort", nil, "random", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &dto.ListConversationsRequest{
				State: tt.state,
				Sort:  tt.sort,
			}
			errs := req.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation error")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
		})
	}
}

func TestSearchConversationsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{"valid phone", "+1234567890", false},
		{"empty phone", "", true},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &dto.SearchConversationsRequest{Phone: tt.phone}
			errs := req.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation error")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
		})
	}
}

func TestSearchConversationsRequest_NormalizedPhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+1 234 567 890", "+1234567890"},
		{"+1-234-567-890", "+1234567890"},
		{"  +1234567890  ", "+1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			req := &dto.SearchConversationsRequest{Phone: tt.input}
			if got := req.NormalizedPhone(); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseListConversationsRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/conversations?state=QUEUED&sort=priority&per_page=25", nil)
	parsed := dto.ParseListConversationsRequest(req)

	if parsed.State == nil || *parsed.State != "QUEUED" {
		t.Error("state not parsed correctly")
	}
	if parsed.Sort != "priority" {
		t.Errorf("sort: got %q, want %q", parsed.Sort, "priority")
	}
	if parsed.PerPage != 25 {
		t.Errorf("per_page: got %d, want 25", parsed.PerPage)
	}
}

func TestParseListConversationsRequest_Defaults(t *testing.T) {
	req := httptest.NewRequest("GET", "/conversations", nil)
	parsed := dto.ParseListConversationsRequest(req)

	if parsed.State != nil {
		t.Error("state should be nil by default")
	}
	if parsed.Sort != dto.SortNewest {
		t.Errorf("sort: got %q, want %q", parsed.Sort, dto.SortNewest)
	}
	if parsed.PerPage != dto.DefaultPerPage {
		t.Errorf("per_page: got %d, want %d", parsed.PerPage, dto.DefaultPerPage)
	}
}

func TestParseListConversationsRequest_MaxPerPage(t *testing.T) {
	req := httptest.NewRequest("GET", "/conversations?per_page=500", nil)
	parsed := dto.ParseListConversationsRequest(req)

	if parsed.PerPage > dto.MaxConversationsPerQuery {
		t.Errorf("per_page should be capped at %d, got %d", dto.MaxConversationsPerQuery, parsed.PerPage)
	}
}

// Helper
func strPtr(s string) *string {
	return &s
}
