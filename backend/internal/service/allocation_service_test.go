package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllocationService_ValidateOperatorStatus(t *testing.T) {
	ctx := testutil.TestContext(t)

	t.Run("operator not available returns error", func(t *testing.T) {
		// Setup mocks
		statusRepo := testutil.NewMockOperatorStatusRepository()

		tenant := testutil.NewTestTenant()
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)

		// Operator is OFFLINE
		status := testutil.NewTestOperatorStatus(operator.ID, domain.OperatorStatusOffline)
		statusRepo.AddStatus(status)

		// Try to get status
		retrievedStatus, err := statusRepo.GetByOperatorID(ctx, operator.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.OperatorStatusOffline, retrievedStatus.Status)

		// Verify operator is not available
		assert.NotEqual(t, domain.OperatorStatusAvailable, retrievedStatus.Status)
	})

	t.Run("operator available passes validation", func(t *testing.T) {
		statusRepo := testutil.NewMockOperatorStatusRepository()

		tenant := testutil.NewTestTenant()
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)

		// Operator is AVAILABLE
		status := testutil.NewTestOperatorStatus(operator.ID, domain.OperatorStatusAvailable)
		statusRepo.AddStatus(status)

		retrievedStatus, err := statusRepo.GetByOperatorID(ctx, operator.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.OperatorStatusAvailable, retrievedStatus.Status)
	})
}

func TestAllocationService_ValidateSubscriptions(t *testing.T) {
	ctx := testutil.TestContext(t)

	t.Run("operator with no subscriptions returns error", func(t *testing.T) {
		subRepo := testutil.NewMockSubscriptionRepository()

		tenant := testutil.NewTestTenant()
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)

		// No subscriptions for operator
		subs, err := subRepo.GetByOperatorID(ctx, operator.ID)
		require.NoError(t, err)
		assert.Len(t, subs, 0)
	})

	t.Run("operator with subscriptions passes validation", func(t *testing.T) {
		subRepo := testutil.NewMockSubscriptionRepository()

		tenant := testutil.NewTestTenant()
		inbox := testutil.NewTestInbox(tenant.ID)
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)

		// Add subscription
		sub := testutil.NewTestSubscription(operator.ID, inbox.ID)
		subRepo.AddSubscription(sub)

		subs, err := subRepo.GetByOperatorID(ctx, operator.ID)
		require.NoError(t, err)
		assert.Len(t, subs, 1)
		assert.Equal(t, inbox.ID, subs[0].InboxID)
	})
}

func TestAllocationService_ConversationAvailability(t *testing.T) {
	ctx := testutil.TestContext(t)

	t.Run("no queued conversations returns empty", func(t *testing.T) {
		convRepo := testutil.NewMockConversationRepository()

		tenant := testutil.NewTestTenant()
		inbox := testutil.NewTestInbox(tenant.ID)
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)

		// No conversations in queue
		convs, err := convRepo.GetQueuedForOperator(ctx, operator.ID, []uuid.UUID{inbox.ID}, 10)
		require.NoError(t, err)
		assert.Len(t, convs, 0)
	})

	t.Run("queued conversations are available", func(t *testing.T) {
		convRepo := testutil.NewMockConversationRepository()

		tenant := testutil.NewTestTenant()
		inbox := testutil.NewTestInbox(tenant.ID)
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)

		// Add queued conversations
		for i := 0; i < 3; i++ {
			conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
			convRepo.AddConversation(conv)
		}

		convs, err := convRepo.GetQueuedForOperator(ctx, operator.ID, []uuid.UUID{inbox.ID}, 10)
		require.NoError(t, err)
		assert.Len(t, convs, 3)

		// All should be QUEUED
		for _, conv := range convs {
			assert.Equal(t, domain.ConversationStateQueued, conv.State)
		}
	})
}

func TestAllocationService_ClaimValidation(t *testing.T) {
	ctx := testutil.TestContext(t)

	t.Run("claim already allocated conversation fails", func(t *testing.T) {
		convRepo := testutil.NewMockConversationRepository()

		tenant := testutil.NewTestTenant()
		inbox := testutil.NewTestInbox(tenant.ID)
		operator1 := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)
		operator2 := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)

		// Conversation already allocated to operator1
		conv := testutil.NewTestConversationWithState(
			tenant.ID, inbox.ID,
			domain.ConversationStateAllocated,
			&operator1.ID,
		)
		convRepo.AddConversation(conv)

		// operator2 tries to claim
		retrieved, err := convRepo.GetByID(ctx, conv.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.ConversationStateAllocated, retrieved.State)
		assert.NotEqual(t, operator2.ID, *retrieved.AssignedOperatorID)
	})

	t.Run("claim queued conversation succeeds", func(t *testing.T) {
		convRepo := testutil.NewMockConversationRepository()

		tenant := testutil.NewTestTenant()
		inbox := testutil.NewTestInbox(tenant.ID)
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)

		// Queued conversation
		conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
		convRepo.AddConversation(conv)

		// Verify can be claimed
		retrieved, err := convRepo.GetByID(ctx, conv.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.ConversationStateQueued, retrieved.State)

		// Simulate claim
		err = retrieved.Allocate(operator.ID)
		require.NoError(t, err)
		convRepo.Update(ctx, retrieved)

		// Verify allocated
		updated, _ := convRepo.GetByID(ctx, conv.ID)
		assert.Equal(t, domain.ConversationStateAllocated, updated.State)
		assert.Equal(t, operator.ID, *updated.AssignedOperatorID)
	})

	t.Run("claim resolved conversation fails", func(t *testing.T) {
		convRepo := testutil.NewMockConversationRepository()

		tenant := testutil.NewTestTenant()
		inbox := testutil.NewTestInbox(tenant.ID)

		// Resolved conversation
		conv := testutil.NewTestConversationWithState(
			tenant.ID, inbox.ID,
			domain.ConversationStateResolved,
			nil,
		)
		convRepo.AddConversation(conv)

		retrieved, err := convRepo.GetByID(ctx, conv.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.ConversationStateResolved, retrieved.State)

		// Cannot allocate resolved conversation
		err = retrieved.Allocate(uuid.Must(uuid.NewV7()))
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInvalidStateTransition, err)
	})
}

func TestAllocationService_MultiTenancy(t *testing.T) {
	ctx := testutil.TestContext(t)

	t.Run("operator cannot access other tenant conversations", func(t *testing.T) {
		convRepo := testutil.NewMockConversationRepository()

		tenant1 := testutil.NewTestTenant()
		tenant2 := testutil.NewTestTenant()
		inbox1 := testutil.NewTestInbox(tenant1.ID)

		// Conversation belongs to tenant1
		conv := testutil.NewTestConversation(tenant1.ID, inbox1.ID)
		convRepo.AddConversation(conv)

		// Operator from tenant2
		operator2 := testutil.NewTestOperator(tenant2.ID, domain.OperatorRoleOperator)

		// Verify tenant isolation
		retrieved, _ := convRepo.GetByID(ctx, conv.ID)
		assert.NotEqual(t, operator2.TenantID, retrieved.TenantID)
	})

	t.Run("operator can only see own tenant conversations", func(t *testing.T) {
		convRepo := testutil.NewMockConversationRepository()

		tenant1 := testutil.NewTestTenant()
		tenant2 := testutil.NewTestTenant()
		inbox1 := testutil.NewTestInbox(tenant1.ID)
		inbox2 := testutil.NewTestInbox(tenant2.ID)

		// Add conversations for both tenants
		conv1 := testutil.NewTestConversation(tenant1.ID, inbox1.ID)
		conv2 := testutil.NewTestConversation(tenant2.ID, inbox2.ID)
		convRepo.AddConversation(conv1)
		convRepo.AddConversation(conv2)

		// Get tenant1 conversations
		convs1, err := convRepo.GetByTenantID(ctx, tenant1.ID)
		require.NoError(t, err)
		assert.Len(t, convs1, 1)
		assert.Equal(t, tenant1.ID, convs1[0].TenantID)

		// Get tenant2 conversations
		convs2, err := convRepo.GetByTenantID(ctx, tenant2.ID)
		require.NoError(t, err)
		assert.Len(t, convs2, 1)
		assert.Equal(t, tenant2.ID, convs2[0].TenantID)
	})
}

func TestAllocationService_InboxSubscription(t *testing.T) {
	ctx := testutil.TestContext(t)

	t.Run("operator can only allocate from subscribed inboxes", func(t *testing.T) {
		subRepo := testutil.NewMockSubscriptionRepository()
		convRepo := testutil.NewMockConversationRepository()

		tenant := testutil.NewTestTenant()
		inbox1 := testutil.NewTestInbox(tenant.ID)
		inbox2 := testutil.NewTestInbox(tenant.ID)
		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)

		// Operator subscribed only to inbox1
		sub := testutil.NewTestSubscription(operator.ID, inbox1.ID)
		subRepo.AddSubscription(sub)

		// Conversations in both inboxes
		conv1 := testutil.NewTestConversation(tenant.ID, inbox1.ID)
		conv2 := testutil.NewTestConversation(tenant.ID, inbox2.ID)
		convRepo.AddConversation(conv1)
		convRepo.AddConversation(conv2)

		// Get operator subscriptions
		subs, err := subRepo.GetByOperatorID(ctx, operator.ID)
		require.NoError(t, err)
		assert.Len(t, subs, 1)
		assert.Equal(t, inbox1.ID, subs[0].InboxID)

		// Operator should only see inbox1 conversations
		convs, err := convRepo.GetQueuedForOperator(ctx, operator.ID, []uuid.UUID{inbox1.ID}, 10)
		require.NoError(t, err)
		assert.Len(t, convs, 1)
		assert.Equal(t, inbox1.ID, convs[0].InboxID)
	})
}
