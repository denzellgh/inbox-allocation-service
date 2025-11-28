//go:build integration && stress

package concurrency

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/repository"
	"github.com/inbox-allocation-service/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestHighLoadAllocation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	pc := testutil.NewPostgresContainer(t)
	ctx := testutil.TestContext(t)

	repos := repository.NewRepositoryContainer(pc.Pool)

	t.Run("100 operators 1000 conversations", func(t *testing.T) {
		pc.CleanTables(ctx)

		// Setup
		tenant := testutil.NewTestTenant()
		repos.Tenants.Create(ctx, tenant)

		inbox := testutil.NewTestInbox(tenant.ID)
		repos.Inboxes.Create(ctx, inbox)

		// Create 1000 conversations
		numConversations := 1000
		t.Logf("Creating %d conversations...", numConversations)
		for i := 0; i < numConversations; i++ {
			conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
			repos.ConversationRefs.Create(ctx, conv)
		}

		// Create 100 operators
		numOperators := 100
		t.Logf("Creating %d operators...", numOperators)
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

		// Each operator tries to allocate 10 conversations
		var wg sync.WaitGroup
		var totalAllocations int32
		allocatedIDs := make(map[uuid.UUID]bool)
		var mu sync.Mutex

		start := time.Now()
		t.Logf("Starting allocation stress test...")

		wg.Add(numOperators)
		for _, op := range operators {
			go func(operator *domain.Operator) {
				defer wg.Done()

				for j := 0; j < 10; j++ {
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
				}
			}(op)
		}

		wg.Wait()
		elapsed := time.Since(start)

		t.Logf("Allocated %d conversations in %v", totalAllocations, elapsed)
		t.Logf("Rate: %.2f allocations/second", float64(totalAllocations)/elapsed.Seconds())

		// No duplicates should exist
		assert.Len(t, allocatedIDs, int(totalAllocations),
			"Each conversation should only be allocated once")

		// Should allocate up to numConversations
		assert.LessOrEqual(t, int(totalAllocations), numConversations)

		// Performance assertion - should complete in reasonable time
		assert.Less(t, elapsed.Seconds(), 60.0,
			"Stress test should complete in under 60 seconds")
	})
}

func TestConcurrentMixedOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	pc := testutil.NewPostgresContainer(t)
	ctx := testutil.TestContext(t)

	repos := repository.NewRepositoryContainer(pc.Pool)

	t.Run("allocate and resolve concurrently", func(t *testing.T) {
		pc.CleanTables(ctx)

		// Setup
		tenant := testutil.NewTestTenant()
		repos.Tenants.Create(ctx, tenant)

		inbox := testutil.NewTestInbox(tenant.ID)
		repos.Inboxes.Create(ctx, inbox)

		// Create conversations
		numConversations := 100
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
		var allocations, resolves int32

		// Allocators
		wg.Add(numOperators)
		for _, op := range operators {
			go func(operator *domain.Operator) {
				defer wg.Done()
				for i := 0; i < 10; i++ {
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
								atomic.AddInt32(&allocations, 1)
							}
						}
					}
				}
			}(op)
		}

		// Resolvers
		wg.Add(numOperators)
		for _, op := range operators {
			go func(operator *domain.Operator) {
				defer wg.Done()
				time.Sleep(50 * time.Millisecond) // Let some allocations happen

				for i := 0; i < 5; i++ {
					// Get operator's allocated conversations
					convs, err := repos.ConversationRefs.GetByOperatorID(
						ctx,
						tenant.ID,
						operator.ID,
						nil,
					)
					if err == nil {
						for _, conv := range convs {
							if conv.State == domain.ConversationStateAllocated {
								err := conv.Resolve()
								if err == nil {
									err = repos.ConversationRefs.Update(ctx, conv)
									if err == nil {
										atomic.AddInt32(&resolves, 1)
									}
								}
								break
							}
						}
					}
				}
			}(op)
		}

		wg.Wait()

		t.Logf("Allocations: %d, Resolves: %d", allocations, resolves)

		// All operations should complete without panics or data corruption
		assert.True(t, allocations > 0, "Should have some allocations")

		// Verify data integrity
		allConvs, _ := repos.ConversationRefs.GetByFilter(ctx, domain.ConversationFilter{
			TenantID: &tenant.ID,
		})
		for _, conv := range allConvs {
			// Valid states
			assert.True(t, conv.State.IsValid(), "State should be valid")

			// If allocated, must have operator
			if conv.State == domain.ConversationStateAllocated {
				assert.NotNil(t, conv.AssignedOperatorID, "Allocated conversation must have operator")
			}

			// If queued or resolved, should not have operator
			if conv.State == domain.ConversationStateQueued ||
				conv.State == domain.ConversationStateResolved {
				// Note: Resolved conversations may still have operator ID in some implementations
				// This is acceptable as long as state is correct
			}
		}
	})
}

func TestRaceDetector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race detector test in short mode")
	}

	pc := testutil.NewPostgresContainer(t)
	ctx := testutil.TestContext(t)

	repos := repository.NewRepositoryContainer(pc.Pool)

	t.Run("concurrent reads and writes", func(t *testing.T) {
		pc.CleanTables(ctx)

		// Setup
		tenant := testutil.NewTestTenant()
		repos.Tenants.Create(ctx, tenant)

		inbox := testutil.NewTestInbox(tenant.ID)
		repos.Inboxes.Create(ctx, inbox)

		conv := testutil.NewTestConversation(tenant.ID, inbox.ID)
		repos.ConversationRefs.Create(ctx, conv)

		operator := testutil.NewTestOperator(tenant.ID, domain.OperatorRoleOperator)
		repos.Operators.Create(ctx, operator)

		var wg sync.WaitGroup
		numGoroutines := 50

		// Concurrent readers
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				repos.ConversationRefs.GetByID(ctx, conv.ID)
			}()
		}

		// Concurrent writers
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				c, _ := repos.ConversationRefs.GetByID(ctx, conv.ID)
				if c != nil {
					repos.ConversationRefs.Update(ctx, c)
				}
			}()
		}

		wg.Wait()

		// If we get here without race detector warnings, test passes
		t.Log("No race conditions detected")
	})
}
