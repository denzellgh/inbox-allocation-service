package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/shopspring/decimal"
)

// TestContext returns a context with timeout for tests
func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// ==================== Fixtures ====================

// NewTestTenant creates a tenant for testing
func NewTestTenant() *domain.Tenant {
	return domain.NewTenant(
		"Test Tenant",
		decimal.NewFromFloat(0.6),
		decimal.NewFromFloat(0.4),
	)
}

// NewTestInbox creates an inbox for testing
func NewTestInbox(tenantID uuid.UUID) *domain.Inbox {
	return domain.NewInbox(
		tenantID,
		"+1234567890",
		"Test Inbox",
	)
}

// NewTestOperator creates an operator for testing
func NewTestOperator(tenantID uuid.UUID, role domain.OperatorRole) *domain.Operator {
	return domain.NewOperator(tenantID, role)
}

// NewTestOperatorWithID creates an operator with specific ID
func NewTestOperatorWithID(id, tenantID uuid.UUID, role domain.OperatorRole) *domain.Operator {
	op := domain.NewOperator(tenantID, role)
	op.ID = id
	return op
}

// NewTestConversation creates a conversation for testing
func NewTestConversation(tenantID, inboxID uuid.UUID) *domain.ConversationRef {
	return domain.NewConversationRef(
		tenantID,
		inboxID,
		uuid.Must(uuid.NewV7()).String(),
		"+1987654321", // Customer phone
	)
}

// NewTestConversationWithState creates a conversation with specific state
func NewTestConversationWithState(
	tenantID, inboxID uuid.UUID,
	state domain.ConversationState,
	operatorID *uuid.UUID,
) *domain.ConversationRef {
	conv := NewTestConversation(tenantID, inboxID)
	conv.State = state
	conv.AssignedOperatorID = operatorID
	return conv
}

// NewTestLabel creates a label for testing
func NewTestLabel(tenantID, inboxID uuid.UUID) *domain.Label {
	color := "#FF0000"
	return domain.NewLabel(tenantID, inboxID, "test-label", &color, nil)
}

// NewTestSubscription creates a subscription for testing
func NewTestSubscription(operatorID, inboxID uuid.UUID) *domain.OperatorInboxSubscription {
	return domain.NewOperatorInboxSubscription(operatorID, inboxID)
}

// NewTestOperatorStatus creates an operator status for testing
func NewTestOperatorStatus(operatorID uuid.UUID, status domain.OperatorStatusType) *domain.OperatorStatus {
	opStatus := domain.NewOperatorStatus(operatorID)
	opStatus.SetStatus(status)
	return opStatus
}

// NewTestGracePeriod creates a grace period for testing
func NewTestGracePeriod(conversationID, operatorID uuid.UUID, expiresAt time.Time) *domain.GracePeriodAssignment {
	return domain.NewGracePeriodAssignment(
		conversationID,
		operatorID,
		expiresAt,
		domain.GracePeriodReasonOffline,
	)
}

// ==================== Helpers ====================

// UUIDPtr returns a pointer to a UUID
func UUIDPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

// TimePtr returns a pointer to a time
func TimePtr(t time.Time) *time.Time {
	return &t
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// Int32Ptr returns a pointer to an int32
func Int32Ptr(i int32) *int32 {
	return &i
}
