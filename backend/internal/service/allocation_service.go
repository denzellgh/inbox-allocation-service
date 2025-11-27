package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	ErrOperatorNotAvailable       = errors.New("operator is not available")
	ErrNoSubscriptions            = errors.New("operator has no inbox subscriptions")
	ErrNoConversationsAvailable   = errors.New("no conversations available for allocation")
	ErrConversationNotQueued      = errors.New("conversation is not in QUEUED state")
	ErrConversationAlreadyClaimed = errors.New("conversation has already been claimed")
	ErrNotSubscribedToInbox       = errors.New("operator is not subscribed to this inbox")
)

const MaxAllocationCandidates = 100

type AllocationService struct {
	repos  *repository.RepositoryContainer
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewAllocationService(repos *repository.RepositoryContainer, pool *pgxpool.Pool, logger *zap.Logger) *AllocationService {
	return &AllocationService{
		repos:  repos,
		pool:   pool,
		logger: logger,
	}
}

// ==================== Allocate ====================

// Allocate automatically assigns the next highest-priority conversation to the operator
// CRITICAL: Uses FOR UPDATE SKIP LOCKED to prevent race conditions
func (s *AllocationService) Allocate(ctx context.Context, tenantID, operatorID uuid.UUID) (*domain.ConversationRef, error) {
	start := time.Now()

	// 1. Validate operator status
	status, err := s.repos.OperatorStatus.GetByOperatorID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if status.Status != domain.OperatorStatusAvailable {
		s.logger.Warn("Allocation attempt by non-available operator",
			zap.String("operator_id", operatorID.String()),
			zap.String("status", string(status.Status)))
		return nil, ErrOperatorNotAvailable
	}

	// 2. Get operator's subscribed inboxes
	inboxIDs, err := s.repos.Subscriptions.GetSubscribedInboxIDs(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if len(inboxIDs) == 0 {
		s.logger.Warn("Allocation attempt by operator with no subscriptions",
			zap.String("operator_id", operatorID.String()))
		return nil, ErrNoSubscriptions
	}

	// 3. Begin transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 4. Get next conversation with lock (FOR UPDATE SKIP LOCKED)
	// This query is CRITICAL for preventing race conditions
	conversations, err := s.repos.ConversationRefs.GetNextForAllocation(ctx, tenantID, inboxIDs, 1)
	if err != nil {
		s.logger.Error("Failed to fetch conversations for allocation",
			zap.String("tenant_id", tenantID.String()),
			zap.String("operator_id", operatorID.String()),
			zap.Error(err))
		return nil, err
	}

	if len(conversations) == 0 {
		s.logger.Debug("No conversations available for allocation",
			zap.String("tenant_id", tenantID.String()),
			zap.String("operator_id", operatorID.String()),
			zap.Strings("inbox_ids", uuidSliceToStringSlice(inboxIDs)))
		return nil, ErrNoConversationsAvailable
	}

	conv := conversations[0]

	// 5. Verify conversation is still QUEUED (should always be true with lock)
	if conv.State != domain.ConversationStateQueued {
		s.logger.Error("Conversation not in QUEUED state after lock",
			zap.String("conversation_id", conv.ID.String()),
			zap.String("state", string(conv.State)))
		return nil, ErrConversationNotQueued
	}

	// 6. Update conversation state to ALLOCATED
	conv.State = domain.ConversationStateAllocated
	conv.AssignedOperatorID = &operatorID
	conv.UpdatedAt = time.Now().UTC()

	if err := s.repos.ConversationRefs.Update(ctx, conv); err != nil {
		s.logger.Error("Failed to update conversation for allocation",
			zap.String("conversation_id", conv.ID.String()),
			zap.Error(err))
		return nil, err
	}

	// 7. Commit transaction
	if err := tx.Commit(ctx); err != nil {
		s.logger.Error("Failed to commit allocation transaction",
			zap.String("conversation_id", conv.ID.String()),
			zap.Error(err))
		return nil, err
	}

	// 8. Log success
	priorityScore, _ := conv.PriorityScore.Float64()
	s.logger.Info("Conversation allocated",
		zap.String("conversation_id", conv.ID.String()),
		zap.String("operator_id", operatorID.String()),
		zap.String("inbox_id", conv.InboxID.String()),
		zap.String("tenant_id", tenantID.String()),
		zap.Float64("priority_score", priorityScore),
		zap.Duration("allocation_time", time.Since(start)))

	return conv, nil
}

// ==================== Claim ====================

// Claim allows an operator to manually claim a specific QUEUED conversation
// CRITICAL: Uses FOR UPDATE NOWAIT to fail fast if conversation is locked
func (s *AllocationService) Claim(ctx context.Context, tenantID, operatorID, conversationID uuid.UUID) (*domain.ConversationRef, error) {
	start := time.Now()

	// 1. Validate operator status
	status, err := s.repos.OperatorStatus.GetByOperatorID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if status.Status != domain.OperatorStatusAvailable {
		s.logger.Warn("Claim attempt by non-available operator",
			zap.String("operator_id", operatorID.String()),
			zap.String("status", string(status.Status)))
		return nil, ErrOperatorNotAvailable
	}

	// 2. Begin transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 3. Lock conversation (FOR UPDATE NOWAIT)
	// This will fail immediately if another transaction has locked the row
	conv, err := s.repos.ConversationRefs.LockForClaim(ctx, conversationID)
	if err != nil {
		// Check if it's a lock acquisition error
		if errors.Is(err, domain.ErrLockTimeout) || errors.Is(err, domain.ErrConversationLocked) {
			s.logger.Warn("Conversation already locked for claim",
				zap.String("conversation_id", conversationID.String()),
				zap.String("operator_id", operatorID.String()))
			return nil, ErrConversationAlreadyClaimed
		}
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		s.logger.Error("Failed to lock conversation for claim",
			zap.String("conversation_id", conversationID.String()),
			zap.Error(err))
		return nil, err
	}

	// 4. Verify tenant
	if conv.TenantID != tenantID {
		s.logger.Warn("Claim attempt for conversation in different tenant",
			zap.String("conversation_id", conversationID.String()),
			zap.String("expected_tenant", tenantID.String()),
			zap.String("actual_tenant", conv.TenantID.String()))
		return nil, domain.ErrNotFound
	}

	// 5. Check if conversation is QUEUED
	if conv.State != domain.ConversationStateQueued {
		// If already allocated to this operator, return success (idempotent)
		if conv.State == domain.ConversationStateAllocated &&
			conv.AssignedOperatorID != nil &&
			*conv.AssignedOperatorID == operatorID {
			s.logger.Debug("Conversation already claimed by same operator",
				zap.String("conversation_id", conversationID.String()),
				zap.String("operator_id", operatorID.String()))
			return conv, nil
		}

		s.logger.Warn("Claim attempt for non-QUEUED conversation",
			zap.String("conversation_id", conversationID.String()),
			zap.String("state", string(conv.State)))
		return nil, ErrConversationNotQueued
	}

	// 6. Verify operator is subscribed to the inbox
	isSubscribed, err := s.repos.Subscriptions.IsSubscribed(ctx, operatorID, conv.InboxID)
	if err != nil {
		return nil, err
	}
	if !isSubscribed {
		s.logger.Warn("Claim attempt for conversation in non-subscribed inbox",
			zap.String("conversation_id", conversationID.String()),
			zap.String("operator_id", operatorID.String()),
			zap.String("inbox_id", conv.InboxID.String()))
		return nil, ErrNotSubscribedToInbox
	}

	// 7. Update conversation state to ALLOCATED
	conv.State = domain.ConversationStateAllocated
	conv.AssignedOperatorID = &operatorID
	conv.UpdatedAt = time.Now().UTC()

	if err := s.repos.ConversationRefs.Update(ctx, conv); err != nil {
		s.logger.Error("Failed to update conversation for claim",
			zap.String("conversation_id", conversationID.String()),
			zap.Error(err))
		return nil, err
	}

	// 8. Commit transaction
	if err := tx.Commit(ctx); err != nil {
		s.logger.Error("Failed to commit claim transaction",
			zap.String("conversation_id", conversationID.String()),
			zap.Error(err))
		return nil, err
	}

	// 9. Log success
	priorityScore, _ := conv.PriorityScore.Float64()
	s.logger.Info("Conversation claimed",
		zap.String("conversation_id", conversationID.String()),
		zap.String("operator_id", operatorID.String()),
		zap.String("inbox_id", conv.InboxID.String()),
		zap.String("tenant_id", tenantID.String()),
		zap.Float64("priority_score", priorityScore),
		zap.Duration("claim_time", time.Since(start)))

	return conv, nil
}

// ==================== Helpers ====================

func uuidSliceToStringSlice(ids []uuid.UUID) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}
