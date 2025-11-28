package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Tenant Tests ====================

func TestNewTenant(t *testing.T) {
	alpha := decimal.NewFromFloat(0.6)
	beta := decimal.NewFromFloat(0.4)

	tenant := NewTenant("Test Tenant", alpha, beta)

	require.NotNil(t, tenant)
	assert.NotEqual(t, uuid.Nil, tenant.ID)
	assert.Equal(t, "Test Tenant", tenant.Name)
	assert.True(t, alpha.Equal(tenant.PriorityWeightAlpha))
	assert.True(t, beta.Equal(tenant.PriorityWeightBeta))
	assert.False(t, tenant.CreatedAt.IsZero())
	assert.False(t, tenant.UpdatedAt.IsZero())
	assert.Nil(t, tenant.UpdatedBy)
}

func TestTenant_Weights_SumToOne(t *testing.T) {
	alpha := decimal.NewFromFloat(0.6)
	beta := decimal.NewFromFloat(0.4)

	tenant := NewTenant("Test", alpha, beta)

	sum := tenant.PriorityWeightAlpha.Add(tenant.PriorityWeightBeta)
	assert.True(t, sum.Equal(decimal.NewFromInt(1)), "Weights should sum to 1")
}

// ==================== Inbox Tests ====================

func TestNewInbox(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV7())

	inbox := NewInbox(tenantID, "+1234567890", "Test Inbox")

	require.NotNil(t, inbox)
	assert.NotEqual(t, uuid.Nil, inbox.ID)
	assert.Equal(t, tenantID, inbox.TenantID)
	assert.Equal(t, "+1234567890", inbox.PhoneNumber)
	assert.Equal(t, "Test Inbox", inbox.DisplayName)
	assert.False(t, inbox.CreatedAt.IsZero())
	assert.False(t, inbox.UpdatedAt.IsZero())
}

// ==================== Operator Tests ====================

func TestNewOperator(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name string
		role OperatorRole
	}{
		{"Operator role", OperatorRoleOperator},
		{"Manager role", OperatorRoleManager},
		{"Admin role", OperatorRoleAdmin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operator := NewOperator(tenantID, tt.role)

			require.NotNil(t, operator)
			assert.NotEqual(t, uuid.Nil, operator.ID)
			assert.Equal(t, tenantID, operator.TenantID)
			assert.Equal(t, tt.role, operator.Role)
			assert.False(t, operator.CreatedAt.IsZero())
		})
	}
}

// ==================== OperatorInboxSubscription Tests ====================

func TestNewOperatorInboxSubscription(t *testing.T) {
	operatorID := uuid.Must(uuid.NewV7())
	inboxID := uuid.Must(uuid.NewV7())

	sub := NewOperatorInboxSubscription(operatorID, inboxID)

	require.NotNil(t, sub)
	assert.NotEqual(t, uuid.Nil, sub.ID)
	assert.Equal(t, operatorID, sub.OperatorID)
	assert.Equal(t, inboxID, sub.InboxID)
	assert.False(t, sub.CreatedAt.IsZero())
}

// ==================== OperatorStatus Tests ====================

func TestNewOperatorStatus(t *testing.T) {
	operatorID := uuid.Must(uuid.NewV7())

	status := NewOperatorStatus(operatorID)

	require.NotNil(t, status)
	assert.NotEqual(t, uuid.Nil, status.ID)
	assert.Equal(t, operatorID, status.OperatorID)
	assert.Equal(t, OperatorStatusOffline, status.Status)
	assert.False(t, status.LastStatusChangeAt.IsZero())
}

func TestOperatorStatus_SetStatus(t *testing.T) {
	operatorID := uuid.Must(uuid.NewV7())
	status := NewOperatorStatus(operatorID)

	originalChangedAt := status.LastStatusChangeAt

	// Wait a tiny bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	status.SetStatus(OperatorStatusAvailable)

	assert.Equal(t, OperatorStatusAvailable, status.Status)
	assert.True(t, status.LastStatusChangeAt.After(originalChangedAt))
}

// ==================== ConversationRef Tests ====================

func TestNewConversationRef(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV7())
	inboxID := uuid.Must(uuid.NewV7())
	externalID := "ext-123"
	customerPhone := "+1234567890"

	conv := NewConversationRef(tenantID, inboxID, externalID, customerPhone)

	require.NotNil(t, conv)
	assert.NotEqual(t, uuid.Nil, conv.ID)
	assert.Equal(t, tenantID, conv.TenantID)
	assert.Equal(t, inboxID, conv.InboxID)
	assert.Equal(t, externalID, conv.ExternalConversationID)
	assert.Equal(t, customerPhone, conv.CustomerPhoneNumber)
	assert.Equal(t, ConversationStateQueued, conv.State)
	assert.Nil(t, conv.AssignedOperatorID)
	assert.Equal(t, int32(0), conv.MessageCount)
	assert.True(t, conv.PriorityScore.IsZero())
	assert.False(t, conv.CreatedAt.IsZero())
}

func TestConversationRef_Allocate(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV7())
	inboxID := uuid.Must(uuid.NewV7())
	operatorID := uuid.Must(uuid.NewV7())

	conv := NewConversationRef(tenantID, inboxID, "ext-1", "+1234567890")

	err := conv.Allocate(operatorID)
	require.NoError(t, err)
	assert.Equal(t, ConversationStateAllocated, conv.State)
	assert.Equal(t, operatorID, *conv.AssignedOperatorID)
}

func TestConversationRef_Deallocate(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV7())
	inboxID := uuid.Must(uuid.NewV7())
	operatorID := uuid.Must(uuid.NewV7())

	conv := NewConversationRef(tenantID, inboxID, "ext-1", "+1234567890")
	conv.Allocate(operatorID)

	err := conv.Deallocate()
	require.NoError(t, err)
	assert.Equal(t, ConversationStateQueued, conv.State)
	assert.Nil(t, conv.AssignedOperatorID)
}

func TestConversationRef_Resolve(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV7())
	inboxID := uuid.Must(uuid.NewV7())
	operatorID := uuid.Must(uuid.NewV7())

	conv := NewConversationRef(tenantID, inboxID, "ext-1", "+1234567890")
	conv.Allocate(operatorID)

	err := conv.Resolve()
	require.NoError(t, err)
	assert.Equal(t, ConversationStateResolved, conv.State)
	assert.NotNil(t, conv.ResolvedAt)
}

// ==================== Label Tests ====================

func TestNewLabel(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV7())
	inboxID := uuid.Must(uuid.NewV7())
	color := "#FF0000"

	label := NewLabel(tenantID, inboxID, "important", &color, nil)

	require.NotNil(t, label)
	assert.NotEqual(t, uuid.Nil, label.ID)
	assert.Equal(t, tenantID, label.TenantID)
	assert.Equal(t, inboxID, label.InboxID)
	assert.Equal(t, "important", label.Name)
	assert.Equal(t, color, *label.Color)
	assert.False(t, label.CreatedAt.IsZero())
}

// ==================== ConversationLabel Tests ====================

func TestNewConversationLabel(t *testing.T) {
	conversationID := uuid.Must(uuid.NewV7())
	labelID := uuid.Must(uuid.NewV7())

	cl := NewConversationLabel(conversationID, labelID)

	require.NotNil(t, cl)
	assert.NotEqual(t, uuid.Nil, cl.ID)
	assert.Equal(t, conversationID, cl.ConversationID)
	assert.Equal(t, labelID, cl.LabelID)
	assert.False(t, cl.CreatedAt.IsZero())
}

// ==================== GracePeriodAssignment Tests ====================

func TestNewGracePeriodAssignment(t *testing.T) {
	conversationID := uuid.Must(uuid.NewV7())
	operatorID := uuid.Must(uuid.NewV7())
	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	gpa := NewGracePeriodAssignment(
		conversationID,
		operatorID,
		expiresAt,
		GracePeriodReasonOffline,
	)

	require.NotNil(t, gpa)
	assert.NotEqual(t, uuid.Nil, gpa.ID)
	assert.Equal(t, conversationID, gpa.ConversationID)
	assert.Equal(t, operatorID, gpa.OperatorID)
	assert.Equal(t, expiresAt.Unix(), gpa.ExpiresAt.Unix())
	assert.Equal(t, GracePeriodReasonOffline, gpa.Reason)
	assert.False(t, gpa.CreatedAt.IsZero())
}

func TestGracePeriodAssignment_IsExpired(t *testing.T) {
	conversationID := uuid.Must(uuid.NewV7())
	operatorID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name      string
		expiresAt time.Time
		expired   bool
	}{
		{"Future expiry is not expired", time.Now().UTC().Add(5 * time.Minute), false},
		{"Past expiry is expired", time.Now().UTC().Add(-5 * time.Minute), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gpa := NewGracePeriodAssignment(
				conversationID,
				operatorID,
				tt.expiresAt,
				GracePeriodReasonOffline,
			)

			assert.Equal(t, tt.expired, gpa.IsExpired())
		})
	}
}

// ==================== IdempotencyKey Tests ====================

func TestNewIdempotencyKey(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV7())
	key := "test-key-123"
	ttl := 24 * time.Hour

	ik := NewIdempotencyKey(
		key,
		tenantID,
		"/api/v1/allocate",
		"POST",
		nil,
		200,
		[]byte(`{"success": true}`),
		ttl,
	)

	require.NotNil(t, ik)
	assert.NotEqual(t, uuid.Nil, ik.ID)
	assert.Equal(t, key, ik.Key)
	assert.Equal(t, tenantID, ik.TenantID)
	assert.Equal(t, "/api/v1/allocate", ik.Endpoint)
	assert.Equal(t, "POST", ik.Method)
	assert.Nil(t, ik.RequestHash)
	assert.Equal(t, 200, ik.ResponseStatus)
	assert.Equal(t, []byte(`{"success": true}`), ik.ResponseBody)
	assert.False(t, ik.CreatedAt.IsZero())
	assert.True(t, ik.ExpiresAt.After(ik.CreatedAt))
}

func TestIdempotencyKey_IsExpired(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name    string
		ttl     time.Duration
		expired bool
	}{
		{"Future expiry is not expired", 24 * time.Hour, false},
		{"Past expiry is expired", -1 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ik := NewIdempotencyKey(
				"key",
				tenantID,
				"/api",
				"POST",
				nil,
				200,
				[]byte("{}"),
				tt.ttl,
			)

			assert.Equal(t, tt.expired, ik.IsExpired())
		})
	}
}
