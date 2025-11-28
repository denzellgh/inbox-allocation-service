package dto_test

import (
	"testing"

	"github.com/inbox-allocation-service/internal/api/dto"
)

func TestCreateInboxRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		displayName string
		wantErrs    int
	}{
		{"valid", "+1234567890", "Support", 0},
		{"missing phone", "", "Support", 1},
		{"missing display", "+1234567890", "", 1},
		{"both missing", "", "", 2},
		{"phone too long", "123456789012345678901", "OK", 1},
		{"display too long", "+123", "Lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua Ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur Excepteur sint occaecat cupidatat non proident sunt in culpa qui officia deserunt mollit anim id est laborum", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dto.CreateInboxRequest{
				PhoneNumber: tt.phoneNumber,
				DisplayName: tt.displayName,
			}
			errs := req.Validate()
			if len(errs) != tt.wantErrs {
				t.Errorf("got %d errors, want %d: %v", len(errs), tt.wantErrs, errs)
			}
		})
	}
}

func TestUpdateInboxRequest_Validate(t *testing.T) {
	phoneOK := "+1234567890"
	phoneTooLong := "123456789012345678901"
	displayOK := "Support"
	displayTooLong := "Lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua Ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur Excepteur sint occaecat cupidatat non proident sunt in culpa qui officia deserunt mollit anim id est laborum"

	tests := []struct {
		name        string
		phoneNumber *string
		displayName *string
		wantErrs    int
	}{
		{"no updates", nil, nil, 0},
		{"valid phone update", &phoneOK, nil, 0},
		{"valid display update", nil, &displayOK, 0},
		{"both valid", &phoneOK, &displayOK, 0},
		{"phone too long", &phoneTooLong, nil, 1},
		{"display too long", nil, &displayTooLong, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dto.UpdateInboxRequest{
				PhoneNumber: tt.phoneNumber,
				DisplayName: tt.displayName,
			}
			errs := req.Validate()
			if len(errs) != tt.wantErrs {
				t.Errorf("got %d errors, want %d: %v", len(errs), tt.wantErrs, errs)
			}
		})
	}
}
