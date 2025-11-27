package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/pkg/logger"
	"github.com/inbox-allocation-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// GracePeriodResult holds the result of processing grace periods
type GracePeriodResult struct {
	Processed      int
	Transitioned   int
	AlreadyHandled int
	Errors         int
}

type GracePeriodService struct {
	repos  *repository.RepositoryContainer
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewGracePeriodService(
	repos *repository.RepositoryContainer,
	pool *pgxpool.Pool,
	log *logger.Logger,
) *GracePeriodService {
	return &GracePeriodService{
		repos:  repos,
		pool:   pool,
		logger: log,
	}
}

// ProcessExpiredGracePeriods processes expired grace period assignments
// Uses FOR UPDATE SKIP LOCKED for distributed processing safety
// Returns the result of processing
func (s *GracePeriodService) ProcessExpiredGracePeriods(ctx context.Context, batchSize int) (*GracePeriodResult, error) {
	start := time.Now()
	result := &GracePeriodResult{}

	// Begin transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get and lock expired grace periods (FOR UPDATE SKIP LOCKED)
	expired, err := s.repos.GracePeriodAssignments.GetAndLockExpired(ctx, batchSize)
	if err != nil {
		return nil, err
	}

	if len(expired) == 0 {
		return result, nil
	}

	result.Processed = len(expired)

	// Process each expired grace period
	for _, gpa := range expired {
		err := s.processGracePeriod(ctx, gpa, result)
		if err != nil {
			s.logger.Error("Failed to process grace period",
				zap.String("grace_period_id", gpa.ID.String()),
				zap.String("conversation_id", gpa.ConversationID.String()),
				zap.Error(err))
			result.Errors++
			continue
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.logger.Info("Grace period processing completed",
		zap.Int("processed", result.Processed),
		zap.Int("transitioned", result.Transitioned),
		zap.Int("already_handled", result.AlreadyHandled),
		zap.Int("errors", result.Errors),
		zap.Duration("duration", time.Since(start)))

	return result, nil
}

// processGracePeriod handles a single grace period expiration
func (s *GracePeriodService) processGracePeriod(
	ctx context.Context,
	gpa *domain.GracePeriodAssignment,
	result *GracePeriodResult,
) error {
	// Get the conversation
	conv, err := s.repos.ConversationRefs.GetByID(ctx, gpa.ConversationID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Conversation was deleted, just remove the grace period
			s.logger.Debug("Conversation not found, removing grace period",
				zap.String("conversation_id", gpa.ConversationID.String()))
			return s.repos.GracePeriodAssignments.Delete(ctx, gpa.ID)
		}
		return err
	}

	// Check if conversation is still ALLOCATED
	if conv.State != domain.ConversationStateAllocated {
		// Already transitioned (resolved, deallocated manually, etc.)
		s.logger.Debug("Conversation already transitioned",
			zap.String("conversation_id", conv.ID.String()),
			zap.String("current_state", string(conv.State)))
		result.AlreadyHandled++
		return s.repos.GracePeriodAssignments.Delete(ctx, gpa.ID)
	}

	// Verify the assigned operator matches (extra safety check)
	if conv.AssignedOperatorID == nil || *conv.AssignedOperatorID != gpa.OperatorID {
		// Reassigned to different operator, just remove grace period
		s.logger.Debug("Conversation reassigned to different operator",
			zap.String("conversation_id", conv.ID.String()),
			zap.String("grace_operator_id", gpa.OperatorID.String()))
		result.AlreadyHandled++
		return s.repos.GracePeriodAssignments.Delete(ctx, gpa.ID)
	}

	// Transition conversation to QUEUED
	if err := conv.Deallocate(); err != nil {
		return err
	}

	if err := s.repos.ConversationRefs.Update(ctx, conv); err != nil {
		return err
	}

	// Delete grace period entry
	if err := s.repos.GracePeriodAssignments.Delete(ctx, gpa.ID); err != nil {
		return err
	}

	s.logger.Info("Conversation returned to queue due to grace period expiration",
		zap.String("conversation_id", conv.ID.String()),
		zap.String("operator_id", gpa.OperatorID.String()),
		zap.String("reason", string(gpa.Reason)))

	result.Transitioned++
	return nil
}

// CreateGracePeriod creates a grace period for a conversation
// Called when manual deallocation with grace is requested
func (s *GracePeriodService) CreateGracePeriod(
	ctx context.Context,
	conversationID, operatorID uuid.UUID,
	duration time.Duration,
	reason domain.GracePeriodReason,
) (*domain.GracePeriodAssignment, error) {
	expiresAt := time.Now().UTC().Add(duration)
	gpa := domain.NewGracePeriodAssignment(conversationID, operatorID, expiresAt, reason)

	if err := s.repos.GracePeriodAssignments.Create(ctx, gpa); err != nil {
		return nil, err
	}

	s.logger.Info("Grace period created",
		zap.String("conversation_id", conversationID.String()),
		zap.String("operator_id", operatorID.String()),
		zap.String("reason", string(reason)),
		zap.Time("expires_at", expiresAt))

	return gpa, nil
}

// GetPendingByOperator returns all pending grace periods for an operator
func (s *GracePeriodService) GetPendingByOperator(
	ctx context.Context,
	operatorID uuid.UUID,
) ([]*domain.GracePeriodAssignment, error) {
	return s.repos.GracePeriodAssignments.GetByOperatorID(ctx, operatorID)
}

// CancelByOperator cancels all grace periods for an operator
// Called when operator returns to AVAILABLE
func (s *GracePeriodService) CancelByOperator(ctx context.Context, operatorID uuid.UUID) error {
	if err := s.repos.GracePeriodAssignments.DeleteByOperatorID(ctx, operatorID); err != nil {
		return err
	}

	s.logger.Info("Grace periods cancelled for operator",
		zap.String("operator_id", operatorID.String()))

	return nil
}

// CancelByConversation cancels grace period for a specific conversation
// Called when conversation is manually resolved or reassigned
func (s *GracePeriodService) CancelByConversation(ctx context.Context, conversationID uuid.UUID) error {
	if err := s.repos.GracePeriodAssignments.DeleteByConversationID(ctx, conversationID); err != nil {
		return err
	}

	s.logger.Debug("Grace period cancelled for conversation",
		zap.String("conversation_id", conversationID.String()))

	return nil
}
