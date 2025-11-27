package dto_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/api/dto"
)

func TestParseAllocateRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "/allocate", nil)
	parsed := dto.ParseAllocateRequest(req)

	if parsed == nil {
		t.Error("expected non-nil request")
	}

	// Should have no validation errors
	if errs := parsed.Validate(); len(errs) > 0 {
		t.Errorf("unexpected validation errors: %v", errs)
	}
}

func TestClaimRequest_Validate(t *testing.T) {
	tests := []struct {
		name           string
		conversationID uuid.UUID
		wantErr        bool
	}{
		{
			name:           "valid UUID",
			conversationID: uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678"),
			wantErr:        false,
		},
		{
			name:           "nil UUID",
			conversationID: uuid.Nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &dto.ClaimRequest{
				ConversationID: tt.conversationID,
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

func TestParseClaimRequest(t *testing.T) {
	validID := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")

	tests := []struct {
		name    string
		body    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid request",
			body: map[string]interface{}{
				"conversation_id": validID.String(),
			},
			wantErr: false,
		},
		{
			name:    "empty body",
			body:    map[string]interface{}{},
			wantErr: false, // Will fail validation, not parsing
		},
		{
			name:    "invalid JSON",
			body:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if tt.body != nil {
				body, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("failed to marshal body: %v", err)
				}
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest("POST", "/claim", bytes.NewReader(body))
			parsed, err := dto.ParseClaimRequest(req)

			if tt.wantErr {
				if err == nil {
					t.Error("expected parsing error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if parsed == nil {
					t.Error("expected non-nil request")
				}
			}
		})
	}
}

func TestClaimRequest_ValidUUID(t *testing.T) {
	validID := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")
	body, _ := json.Marshal(map[string]interface{}{
		"conversation_id": validID.String(),
	})

	req := httptest.NewRequest("POST", "/claim", bytes.NewReader(body))
	parsed, err := dto.ParseClaimRequest(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.ConversationID != validID {
		t.Errorf("expected conversation_id %v, got %v", validID, parsed.ConversationID)
	}

	if errs := parsed.Validate(); len(errs) > 0 {
		t.Errorf("unexpected validation errors: %v", errs)
	}
}
