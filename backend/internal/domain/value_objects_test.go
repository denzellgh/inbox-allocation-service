package domain

import "testing"

func TestConversationState_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		from   ConversationState
		to     ConversationState
		expect bool
	}{
		{"QUEUED to ALLOCATED", ConversationStateQueued, ConversationStateAllocated, true},
		{"QUEUED to RESOLVED", ConversationStateQueued, ConversationStateResolved, false},
		{"ALLOCATED to QUEUED", ConversationStateAllocated, ConversationStateQueued, true},
		{"ALLOCATED to RESOLVED", ConversationStateAllocated, ConversationStateResolved, true},
		{"RESOLVED to QUEUED", ConversationStateResolved, ConversationStateQueued, false},
		{"RESOLVED to ALLOCATED", ConversationStateResolved, ConversationStateAllocated, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.expect {
				t.Errorf("CanTransitionTo() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestConversationState_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		state ConversationState
		want  bool
	}{
		{"QUEUED is valid", ConversationStateQueued, true},
		{"ALLOCATED is valid", ConversationStateAllocated, true},
		{"RESOLVED is valid", ConversationStateResolved, true},
		{"INVALID is not valid", ConversationState("INVALID"), false},
		{"Empty is not valid", ConversationState(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperatorRole_Permissions(t *testing.T) {
	tests := []struct {
		role          OperatorRole
		canDeallocate bool
		canReassign   bool
		canMoveInbox  bool
	}{
		{OperatorRoleOperator, false, false, false},
		{OperatorRoleManager, true, true, true},
		{OperatorRoleAdmin, true, true, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if got := tt.role.CanDeallocate(); got != tt.canDeallocate {
				t.Errorf("CanDeallocate() = %v, want %v", got, tt.canDeallocate)
			}
			if got := tt.role.CanReassign(); got != tt.canReassign {
				t.Errorf("CanReassign() = %v, want %v", got, tt.canReassign)
			}
			if got := tt.role.CanMoveInbox(); got != tt.canMoveInbox {
				t.Errorf("CanMoveInbox() = %v, want %v", got, tt.canMoveInbox)
			}
		})
	}
}

func TestOperatorRole_IsValid(t *testing.T) {
	tests := []struct {
		name string
		role OperatorRole
		want bool
	}{
		{"OPERATOR is valid", OperatorRoleOperator, true},
		{"MANAGER is valid", OperatorRoleManager, true},
		{"ADMIN is valid", OperatorRoleAdmin, true},
		{"INVALID is not valid", OperatorRole("INVALID"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperatorStatusType_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status OperatorStatusType
		want   bool
	}{
		{"AVAILABLE is valid", OperatorStatusAvailable, true},
		{"OFFLINE is valid", OperatorStatusOffline, true},
		{"INVALID is not valid", OperatorStatusType("INVALID"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGracePeriodReason_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		reason GracePeriodReason
		want   bool
	}{
		{"OFFLINE is valid", GracePeriodReasonOffline, true},
		{"MANUAL is valid", GracePeriodReasonManual, true},
		{"INVALID is not valid", GracePeriodReason("INVALID"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reason.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTenantID_IsZero(t *testing.T) {
	tests := []struct {
		name string
		id   TenantID
		want bool
	}{
		{"New ID is not zero", NewTenantID(), false},
		{"Zero ID is zero", TenantID{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTenantID(t *testing.T) {
	validID := NewTenantID()
	validIDStr := validID.String()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid UUID", validIDStr, false},
		{"Invalid UUID", "invalid-uuid", true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTenantID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTenantID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
