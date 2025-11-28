package dto_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/api/dto"
)

func TestCreateLabelRequest_Validate(t *testing.T) {
	validID := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")
	color := "#FF0000"
	longName := string(make([]byte, 65))
	longColor := string(make([]byte, 33))

	tests := []struct {
		name     string
		req      dto.CreateLabelRequest
		wantErr  bool
		errCount int
	}{
		{
			name:     "valid request",
			req:      dto.CreateLabelRequest{InboxID: validID, Name: "Important", Color: &color},
			wantErr:  false,
			errCount: 0,
		},
		{
			name:     "valid without color",
			req:      dto.CreateLabelRequest{InboxID: validID, Name: "Urgent"},
			wantErr:  false,
			errCount: 0,
		},
		{
			name:     "missing inbox_id",
			req:      dto.CreateLabelRequest{Name: "Test"},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "empty name",
			req:      dto.CreateLabelRequest{InboxID: validID, Name: ""},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "whitespace name",
			req:      dto.CreateLabelRequest{InboxID: validID, Name: "   "},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "name too long",
			req:      dto.CreateLabelRequest{InboxID: validID, Name: longName},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "color too long",
			req:      dto.CreateLabelRequest{InboxID: validID, Name: "Test", Color: &longColor},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "multiple errors",
			req:      dto.CreateLabelRequest{},
			wantErr:  true,
			errCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation errors")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if tt.wantErr && len(errs) != tt.errCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errCount, len(errs), errs)
			}
		})
	}
}

func TestUpdateLabelRequest_Validate(t *testing.T) {
	validName := "Updated"
	emptyName := ""
	longName := string(make([]byte, 65))
	color := "#00FF00"
	longColor := string(make([]byte, 33))

	tests := []struct {
		name     string
		req      dto.UpdateLabelRequest
		wantErr  bool
		errCount int
	}{
		{
			name:     "valid name only",
			req:      dto.UpdateLabelRequest{Name: &validName},
			wantErr:  false,
			errCount: 0,
		},
		{
			name:     "valid color only",
			req:      dto.UpdateLabelRequest{Color: &color},
			wantErr:  false,
			errCount: 0,
		},
		{
			name:     "valid both fields",
			req:      dto.UpdateLabelRequest{Name: &validName, Color: &color},
			wantErr:  false,
			errCount: 0,
		},
		{
			name:     "no fields provided",
			req:      dto.UpdateLabelRequest{},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "empty name",
			req:      dto.UpdateLabelRequest{Name: &emptyName},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "name too long",
			req:      dto.UpdateLabelRequest{Name: &longName},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "color too long",
			req:      dto.UpdateLabelRequest{Color: &longColor},
			wantErr:  true,
			errCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation errors")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if tt.wantErr && len(errs) != tt.errCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errCount, len(errs), errs)
			}
		})
	}
}

func TestAttachLabelRequest_Validate(t *testing.T) {
	validID := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")

	tests := []struct {
		name     string
		req      dto.AttachLabelRequest
		wantErr  bool
		errCount int
	}{
		{
			name:     "valid request",
			req:      dto.AttachLabelRequest{ConversationID: validID, LabelID: validID},
			wantErr:  false,
			errCount: 0,
		},
		{
			name:     "missing conversation_id",
			req:      dto.AttachLabelRequest{LabelID: validID},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "missing label_id",
			req:      dto.AttachLabelRequest{ConversationID: validID},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "both missing",
			req:      dto.AttachLabelRequest{},
			wantErr:  true,
			errCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation errors")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if tt.wantErr && len(errs) != tt.errCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errCount, len(errs), errs)
			}
		})
	}
}

func TestDetachLabelRequest_Validate(t *testing.T) {
	validID := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")

	tests := []struct {
		name     string
		req      dto.DetachLabelRequest
		wantErr  bool
		errCount int
	}{
		{
			name:     "valid request",
			req:      dto.DetachLabelRequest{ConversationID: validID, LabelID: validID},
			wantErr:  false,
			errCount: 0,
		},
		{
			name:     "missing conversation_id",
			req:      dto.DetachLabelRequest{LabelID: validID},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "missing label_id",
			req:      dto.DetachLabelRequest{ConversationID: validID},
			wantErr:  true,
			errCount: 1,
		},
		{
			name:     "both missing",
			req:      dto.DetachLabelRequest{},
			wantErr:  true,
			errCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation errors")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if tt.wantErr && len(errs) != tt.errCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errCount, len(errs), errs)
			}
		})
	}
}

func TestParseCreateLabelRequest(t *testing.T) {
	validID := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")
	color := "#FF0000"

	body, _ := json.Marshal(map[string]interface{}{
		"inbox_id": validID.String(),
		"name":     "Important",
		"color":    color,
	})

	req := httptest.NewRequest("POST", "/labels", bytes.NewReader(body))
	parsed, err := dto.ParseCreateLabelRequest(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.InboxID != validID {
		t.Errorf("inbox_id: expected %v, got %v", validID, parsed.InboxID)
	}
	if parsed.Name != "Important" {
		t.Errorf("name: expected Important, got %s", parsed.Name)
	}
	if parsed.Color == nil || *parsed.Color != color {
		t.Errorf("color: expected %s, got %v", color, parsed.Color)
	}
}

func TestParseAttachLabelRequest(t *testing.T) {
	validID := uuid.MustParse("550fc2c9-1234-5678-9abc-def012345678")

	body, _ := json.Marshal(map[string]interface{}{
		"conversation_id": validID.String(),
		"label_id":        validID.String(),
	})

	req := httptest.NewRequest("POST", "/labels/attach", bytes.NewReader(body))
	parsed, err := dto.ParseAttachLabelRequest(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.ConversationID != validID {
		t.Errorf("conversation_id: expected %v, got %v", validID, parsed.ConversationID)
	}
	if parsed.LabelID != validID {
		t.Errorf("label_id: expected %v, got %v", validID, parsed.LabelID)
	}
}
