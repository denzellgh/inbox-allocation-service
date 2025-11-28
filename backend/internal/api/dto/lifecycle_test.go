package dto_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/api/dto"
)

func TestResolveRequest_Validate(t *testing.T) {
	tests := []struct {
		name           string
		conversationID uuid.UUID
		wantErr        bool
	}{
		{"valid UUID", uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678"), false},
		{"nil UUID", uuid.Nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &dto.ResolveRequest{ConversationID: tt.conversationID}
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

func TestDeallocateRequest_Validate(t *testing.T) {
	tests := []struct {
		name           string
		conversationID uuid.UUID
		wantErr        bool
	}{
		{"valid UUID", uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678"), false},
		{"nil UUID", uuid.Nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &dto.DeallocateRequest{ConversationID: tt.conversationID}
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

func TestReassignRequest_Validate(t *testing.T) {
	validID := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")

	tests := []struct {
		name           string
		conversationID uuid.UUID
		operatorID     uuid.UUID
		wantErr        bool
		errCount       int
	}{
		{"all valid", validID, validID, false, 0},
		{"nil conversation", uuid.Nil, validID, true, 1},
		{"nil operator", validID, uuid.Nil, true, 1},
		{"both nil", uuid.Nil, uuid.Nil, true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &dto.ReassignRequest{
				ConversationID: tt.conversationID,
				OperatorID:     tt.operatorID,
			}
			errs := req.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation error")
			}
			if tt.wantErr && len(errs) != tt.errCount {
				t.Errorf("expected %d errors, got %d", tt.errCount, len(errs))
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
		})
	}
}

func TestMoveInboxRequest_Validate(t *testing.T) {
	validID := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")

	tests := []struct {
		name           string
		conversationID uuid.UUID
		inboxID        uuid.UUID
		wantErr        bool
		errCount       int
	}{
		{"all valid", validID, validID, false, 0},
		{"nil conversation", uuid.Nil, validID, true, 1},
		{"nil inbox", validID, uuid.Nil, true, 1},
		{"both nil", uuid.Nil, uuid.Nil, true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &dto.MoveInboxRequest{
				ConversationID: tt.conversationID,
				InboxID:        tt.inboxID,
			}
			errs := req.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation error")
			}
			if tt.wantErr && len(errs) != tt.errCount {
				t.Errorf("expected %d errors, got %d", tt.errCount, len(errs))
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
		})
	}
}

func TestParseResolveRequest(t *testing.T) {
	validID := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")
	body, _ := json.Marshal(map[string]interface{}{
		"conversation_id": validID.String(),
	})

	req := httptest.NewRequest("POST", "/resolve", bytes.NewReader(body))
	parsed, err := dto.ParseResolveRequest(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.ConversationID != validID {
		t.Errorf("expected %v, got %v", validID, parsed.ConversationID)
	}
}

func TestParseReassignRequest(t *testing.T) {
	validID := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")
	body, _ := json.Marshal(map[string]interface{}{
		"conversation_id": validID.String(),
		"operator_id":     validID.String(),
	})

	req := httptest.NewRequest("POST", "/reassign", bytes.NewReader(body))
	parsed, err := dto.ParseReassignRequest(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.ConversationID != validID {
		t.Errorf("conversation_id: expected %v, got %v", validID, parsed.ConversationID)
	}
	if parsed.OperatorID != validID {
		t.Errorf("operator_id: expected %v, got %v", validID, parsed.OperatorID)
	}
}
