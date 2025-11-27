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
	ErrConversationNotAllocated    = errors.New("conversation is not in ALLOCATED state")
	ErrConversationAlreadyResolved = errors.New("conversation is already resolved")
	ErrInsufficientPermissions     = errors.New("insufficient permissions for this operation")
	ErrTargetOperatorNotFound      = errors.New("target operator not found")
	ErrTargetOperatorNotSubscribed = errors.New("target operator is not subscribed to inbox")
	ErrTargetInboxNotFound         = errors.New("target inbox not found")
	ErrTargetInboxDifferentTenant  = errors.New("target inbox belongs to different tenant")
)

type LifecycleService struct {
	repos  *repository.RepositoryContainer
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewLifecycleService(repos *repository.RepositoryContainer, pool *pgxpool.Pool, logger *zap.Logger) *LifecycleService {
	return &LifecycleService{
		repos:  repos,
		pool:   pool,
		logger: logger,
	}
}

// ==================== Resolve ====================

// Resolve marks a conversation as resolved
// Permission: Owner (assigned operator), Manager, or Admin
func (s *LifecycleService) Resolve(ctx context.Context, tenantID, callerID, conversationID uuid.UUID, callerRole domain.OperatorRole) (*domain.ConversationRef, error) {
	start := time.Now()

	// Begin transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get conversation
	conv, err := s.repos.ConversationRefs.GetByID(ctx, conversationID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	// Verify tenant
	if conv.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}

	// Idempotency: if already resolved, return success
	if conv.State == domain.ConversationStateResolved {
		s.logger.Debug("Conversation already resolved",
			zap.String("conversation_id", conversationID.String()))
		return conv, nil
	}

	// Verify state is ALLOCATED
	if conv.State != domain.ConversationStateAllocated {
		return nil, ErrConversationNotAllocated
	}

	// Check permissions
	if !s.canResolve(callerID, callerRole, conv) {
		s.logger.Warn("Resolve attempt without permission",
			zap.String("conversation_id", conversationID.String()),
			zap.String("caller_id", callerID.String()),
			zap.String("caller_role", string(callerRole)))
		return nil, ErrInsufficientPermissions
	}

	// Update state
	now := time.Now().UTC()
	conv.State = domain.ConversationStateResolved
	conv.ResolvedAt = &now
	conv.UpdatedAt = now

	if err := s.repos.ConversationRefs.Update(ctx, conv); err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.logger.Info("Conversation resolved",
		zap.String("conversation_id", conversationID.String()),
		zap.String("resolved_by", callerID.String()),
		zap.String("role", string(callerRole)),
		zap.Duration("duration", time.Since(start)))

	return conv, nil
}

// ==================== Deallocate ====================

// Deallocate returns a conversation to the queue
// Permission: Manager or Admin only
func (s *LifecycleService) Deallocate(ctx context.Context, tenantID, callerID, conversationID uuid.UUID, callerRole domain.OperatorRole) (*domain.ConversationRef, error) {
	start := time.Now()

	// Check permissions first
	if !s.canManage(callerRole) {
		return nil, ErrInsufficientPermissions
	}

	// Begin transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get conversation
	conv, err := s.repos.ConversationRefs.GetByID(ctx, conversationID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	// Verify tenant
	if conv.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}

	// Idempotency: if already queued, return success
	if conv.State == domain.ConversationStateQueued {
		s.logger.Debug("Conversation already queued",
			zap.String("conversation_id", conversationID.String()))
		return conv, nil
	}

	// Verify state is ALLOCATED
	if conv.State != domain.ConversationStateAllocated {
		return nil, ErrConversationNotAllocated
	}

	previousOperator := conv.AssignedOperatorID

	// Update state
	conv.State = domain.ConversationStateQueued
	conv.AssignedOperatorID = nil
	conv.UpdatedAt = time.Now().UTC()

	if err := s.repos.ConversationRefs.Update(ctx, conv); err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	var prevOpStr string
	if previousOperator != nil {
		prevOpStr = previousOperator.String()
	}

	s.logger.Info("Conversation deallocated",
		zap.String("conversation_id", conversationID.String()),
		zap.String("deallocated_by", callerID.String()),
		zap.String("previous_operator", prevOpStr),
		zap.Duration("duration", time.Since(start)))

	return conv, nil
}

// ==================== Reassign ====================

// Reassign assigns a conversation to a different operator
// Permission: Manager or Admin only
func (s *LifecycleService) Reassign(ctx context.Context, tenantID, callerID, conversationID, newOperatorID uuid.UUID, callerRole domain.OperatorRole) (*domain.ConversationRef, error) {
	start := time.Now()

	// Check permissions first
	if !s.canManage(callerRole) {
		return nil, ErrInsufficientPermissions
	}

	// Begin transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get conversation
	conv, err := s.repos.ConversationRefs.GetByID(ctx, conversationID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	// Verify tenant
	if conv.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}

	// Verify state is ALLOCATED
	if conv.State != domain.ConversationStateAllocated {
		return nil, ErrConversationNotAllocated
	}

	// Idempotency: if already assigned to target, return success
	if conv.AssignedOperatorID != nil && *conv.AssignedOperatorID == newOperatorID {
		s.logger.Debug("Conversation already assigned to target operator",
			zap.String("conversation_id", conversationID.String()),
			zap.String("operator_id", newOperatorID.String()))
		return conv, nil
	}

	// Verify new operator exists and is in same tenant
	newOperator, err := s.repos.Operators.GetByID(ctx, newOperatorID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrTargetOperatorNotFound
		}
		return nil, err
	}
	if newOperator.TenantID != tenantID {
		return nil, ErrTargetOperatorNotFound // Don't reveal cross-tenant info
	}

	// Verify new operator is subscribed to the inbox
	isSubscribed, err := s.repos.Subscriptions.IsSubscribed(ctx, newOperatorID, conv.InboxID)
	if err != nil {
		return nil, err
	}
	if !isSubscribed {
		return nil, ErrTargetOperatorNotSubscribed
	}

	previousOperator := conv.AssignedOperatorID

	// Update assignment
	conv.AssignedOperatorID = &newOperatorID
	conv.UpdatedAt = time.Now().UTC()

	if err := s.repos.ConversationRefs.Update(ctx, conv); err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	var prevOpStr string
	if previousOperator != nil {
		prevOpStr = previousOperator.String()
	}

	s.logger.Info("Conversation reassigned",
		zap.String("conversation_id", conversationID.String()),
		zap.String("reassigned_by", callerID.String()),
		zap.String("from_operator", prevOpStr),
		zap.String("to_operator", newOperatorID.String()),
		zap.Duration("duration", time.Since(start)))

	return conv, nil
}

// ==================== Move Inbox ====================

// MoveInbox moves a conversation to a different inbox
// Permission: Manager or Admin only
// Note: If current operator is not subscribed to new inbox, conversation is auto-deallocated
func (s *LifecycleService) MoveInbox(ctx context.Context, tenantID, callerID, conversationID, newInboxID uuid.UUID, callerRole domain.OperatorRole) (*domain.ConversationRef, error) {
	start := time.Now()

	// Check permissions first
	if !s.canManage(callerRole) {
		return nil, ErrInsufficientPermissions
	}

	// Begin transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get conversation
	conv, err := s.repos.ConversationRefs.GetByID(ctx, conversationID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	// Verify tenant
	if conv.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}

	// Idempotency: if already in target inbox, return success
	if conv.InboxID == newInboxID {
		s.logger.Debug("Conversation already in target inbox",
			zap.String("conversation_id", conversationID.String()),
			zap.String("inbox_id", newInboxID.String()))
		return conv, nil
	}

	// Verify new inbox exists and is in same tenant
	newInbox, err := s.repos.Inboxes.GetByID(ctx, newInboxID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrTargetInboxNotFound
		}
		return nil, err
	}
	if newInbox.TenantID != tenantID {
		return nil, ErrTargetInboxDifferentTenant
	}

	previousInbox := conv.InboxID
	autoDeallocated := false

	// If conversation is ALLOCATED, check if operator is subscribed to new inbox
	if conv.State == domain.ConversationStateAllocated && conv.AssignedOperatorID != nil {
		isSubscribed, err := s.repos.Subscriptions.IsSubscribed(ctx, *conv.AssignedOperatorID, newInboxID)
		if err != nil {
			return nil, err
		}
		if !isSubscribed {
			// Auto-deallocate: operator cannot keep conversation in new inbox
			conv.State = domain.ConversationStateQueued
			conv.AssignedOperatorID = nil
			autoDeallocated = true
		}
	}

	// Update inbox
	conv.InboxID = newInboxID
	conv.UpdatedAt = time.Now().UTC()

	if err := s.repos.ConversationRefs.Update(ctx, conv); err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.logger.Info("Conversation moved to new inbox",
		zap.String("conversation_id", conversationID.String()),
		zap.String("moved_by", callerID.String()),
		zap.String("from_inbox", previousInbox.String()),
		zap.String("to_inbox", newInboxID.String()),
		zap.Bool("auto_deallocated", autoDeallocated),
		zap.Duration("duration", time.Since(start)))

	return conv, nil
}

// ==================== Permission Helpers ====================

// canResolve checks if caller can resolve the conversation
func (s *LifecycleService) canResolve(callerID uuid.UUID, callerRole domain.OperatorRole, conv *domain.ConversationRef) bool {
	// Admin and Manager can resolve any conversation
	if callerRole == domain.OperatorRoleAdmin || callerRole == domain.OperatorRoleManager {
		return true
	}
	// Owner can resolve their own conversations
	if conv.AssignedOperatorID != nil && *conv.AssignedOperatorID == callerID {
		return true
	}
	return false
}

// canManage checks if caller has management permissions (Manager or Admin)
func (s *LifecycleService) canManage(callerRole domain.OperatorRole) bool {
	return callerRole == domain.OperatorRoleAdmin || callerRole == domain.OperatorRoleManager
}
