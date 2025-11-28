//go:build integration

package concurrency

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/repository"
	"github.com/inbox-allocation-service/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// TestConcurrentAllocation validates that FOR UPDATE SKIP LOCKED prevents race conditions
func TestConcurrentAllocation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	pc := testutil.NewPostgresContainer(t)
	ctx := testutil.TestContext(t)

	queries := repository.New(pc.Pool)
	repos := repository.NewRepositoryContainer(pc.Pool)

	t.Run("multiple operators compete for same conversation", func(t *testing.T) {
		pc.CleanTables(ctx)

		// Setup
		tenant := testutil.NewTestTenant()
		repos.Tenants.Create(ctx, tenant)

		inbox := testutil.NewTestInbox(tenant.ID)
		repos.Inboxes.Create(ctx, inbox)

		// Create only ONE queued conversation
		conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
		repos.ConversationRefs.Create(ctx, conv)

		// Create 10 operators competing
		numOperators := 10
		operators := make([]*domain.Operator, numOperators)
		for i := 0; i < numOperators; i++ {
			op := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)
			repos.Operators.Create(ctx, op)

			sub := testutil.NewTestSubscription(op.ID, inbox.ID)
			repos.Subscriptions.Create(ctx, sub)

			status := testutil.NewTestOperatorStatus(op.ID, domain.OperatorStatusAvailable)
			repos.OperatorStatus.Create(ctx, status)

			operators[i] = op
		}

		// Concurrent allocation attempts
		var wg sync.WaitGroup
		var successCount int32
		var winnerID uuid.UUID
		var mu sync.Mutex

		wg.Add(numOperators)
		for _, op := range operators {
			go func(operator *domain.Operator) {
				defer wg.Done()

				// Try to get next conversation with lock
				convs, err := repos.ConversationRefs.GetNextForAllocation(
					ctx,
					tenant.ID,
					[]uuid.UUID{inbox.ID},
					1,
				)

				if err == nil && len(convs) > 0 {
					// Try to allocate
					conv := convs[0]
					err := conv.Allocate(operator.ID)
					if err == nil {
						err = repos.ConversationRefs.Update(ctx, conv)
						if err == nil {
							atomic.AddInt32(&successCount, 1)
							mu.Lock()
							winnerID = operator.ID
							mu.Unlock()
						}
					}
				}
			}(op)
		}

		wg.Wait()

		// Only ONE operator should have succeeded
		assert.Equal(t, int32(1), successCount, "Exactly one allocation should succeed")

		// Verify conversation is allocated to exactly one operator
		updated, _ := repos.ConversationRefs.GetByID(ctx, conv.ID)
		assert.Equal(t, domain.ConversationStateAllocated, updated.State)
		assert.NotNil(t, updated.AssignedOperatorID)
		assert.Equal(t, winnerID, *updated.AssignedOperatorID)
	})

	t.Run("FOR UPDATE SKIP LOCKED prevents double allocation", func(t *testing.T) {
		pc.CleanTables(ctx)

		// Setup
		tenant := testutil.NewTestTenant()
		repos.Tenants.Create(ctx, tenant)

		inbox := testutil.NewTestInbox(tenant.ID)
		repos.Inboxes.Create(ctx, inbox)

		// Create multiple queued conversations
		numConversations := 5
		for i := 0; i < numConversations; i++ {
			conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
			repos.ConversationRefs.Create(ctx, conv)
		}

		// Create operators
		numOperators := 10
		operators := make([]*domain.Operator, numOperators)
		for i := 0; i < numOperators; i++ {
			op := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)
			repos.Operators.Create(ctx, op)

			sub := testutil.NewTestSubscription(op.ID, inbox.ID)
			repos.Subscriptions.Create(ctx, sub)

			status := testutil.NewTestOperatorStatus(op.ID, domain.OperatorStatusAvailable)
			repos.OperatorStatus.Create(ctx, status)

			operators[i] = op
		}

		var wg sync.WaitGroup
		allocatedIDs := make(map[uuid.UUID]bool)
		var mu sync.Mutex
		var totalAllocations int32

		wg.Add(numOperators)
		for _, op := range operators {
			go func(operator *domain.Operator) {
				defer wg.Done()

				// Each operator tries to allocate
				convs, err := repos.ConversationRefs.GetNextForAllocation(
					ctx,
					tenant.ID,
					[]uuid.UUID{inbox.ID},
					1,
				)

				if err == nil && len(convs) > 0 {
					conv := convs[0]
					err := conv.Allocate(operator.ID)
					if err == nil {
						err = repos.ConversationRefs.Update(ctx, conv)
						if err == nil {
							atomic.AddInt32(&totalAllocations, 1)

							mu.Lock()
							if allocatedIDs[conv.ID] {
								t.Errorf("Conversation %s was allocated twice!", conv.ID)
							}
							allocatedIDs[conv.ID] = true
							mu.Unlock()
						}
					}
				}
			}(op)
		}

		wg.Wait()

		// Should have allocated exactly numConversations (or less if operators < conversations)
		expectedAllocations := min(numOperators, numConversations)
		assert.LessOrEqual(t, int(totalAllocations), expectedAllocations,
			"Should not allocate more than available conversations")

		// No conversation should be allocated twice
		assert.Len(t, allocatedIDs, int(totalAllocations),
			"Each conversation should only be allocated once")
	})

	_ = queries // silence unused warning
}

func TestConcurrentClaim(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	pc := testutil.NewPostgresContainer(t)
	ctx := testutil.TestContext(t)

	repos := repository.NewRepositoryContainer(pc.Pool)

	t.Run("multiple operators claim same conversation", func(t *testing.T) {
		pc.CleanTables(ctx)

		// Setup
		tenant := testutil.NewTestTenant()
		repos.Tenants.Create(ctx, tenant)

		inbox := testutil.NewTestInbox(tenant.ID)
		repos.Inboxes.Create(ctx, inbox)

		// One conversation
		conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
		repos.ConversationRefs.Create(ctx, conv)

		// Multiple operators
		numOperators := 5
		operators := make([]*domain.Operator, numOperators)
		for i := 0; i < numOperators; i++ {
			op := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)
			repos.Operators.Create(ctx, op)

			sub := testutil.NewTestSubscription(op.ID, inbox.ID)
			repos.Subscriptions.Create(ctx, sub)

			status := testutil.NewTestOperatorStatus(op.ID, domain.OperatorStatusAvailable)
			repos.OperatorStatus.Create(ctx, status)

			operators[i] = op
		}

		var wg sync.WaitGroup
		var successCount int32
		var winnerID uuid.UUID
		var mu sync.Mutex

		wg.Add(numOperators)
		for _, op := range operators {
			go func(operator *domain.Operator) {
				defer wg.Done()

				// Try to lock and claim
				locked, err := repos.ConversationRefs.LockForClaim(ctx, conv.ID)
				if err == nil && locked != nil {
					if locked.State == domain.ConversationStateQueued {
						err := locked.Allocate(operator.ID)
						if err == nil {
							err = repos.ConversationRefs.Update(ctx, locked)
							if err == nil {
								atomic.AddInt32(&successCount, 1)
								mu.Lock()
								winnerID = operator.ID
								mu.Unlock()
							}
						}
					}
				}
			}(op)
		}

		wg.Wait()

		// Only ONE operator should win
		assert.Equal(t, int32(1), successCount, "Exactly one claim should succeed")

		// Verify final state
		updated, _ := repos.ConversationRefs.GetByID(ctx, conv.ID)
		assert.Equal(t, domain.ConversationStateAllocated, updated.State)
		assert.Equal(t, winnerID, *updated.AssignedOperatorID)
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
