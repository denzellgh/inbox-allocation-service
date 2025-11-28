package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/pkg/database"
	"github.com/inbox-allocation-service/internal/pkg/logger"
	"github.com/inbox-allocation-service/internal/repository"
	"go.uber.org/zap"
)

const GracePeriodDuration = 5 * time.Minute

type OperatorService struct {
	repos  *repository.RepositoryContainer
	txMgr  *database.TxManager
	logger *logger.Logger
}

func NewOperatorService(
	repos *repository.RepositoryContainer,
	txMgr *database.TxManager,
	log *logger.Logger,
) *OperatorService {
	return &OperatorService{repos: repos, txMgr: txMgr, logger: log}
}

// ==================== Status Management ====================

func (s *OperatorService) GetStatus(ctx context.Context, operatorID uuid.UUID) (*domain.OperatorStatus, error) {
	return s.repos.OperatorStatus.GetByOperatorID(ctx, operatorID)
}

func (s *OperatorService) UpdateStatus(ctx context.Context, operatorID uuid.UUID, newStatus domain.OperatorStatusType) (*domain.OperatorStatus, error) {
	status, err := s.repos.OperatorStatus.GetByOperatorID(ctx, operatorID)
	if err != nil {
		if err == domain.ErrNotFound {
			// Create initial status
			status = domain.NewOperatorStatus(operatorID)
			status.SetStatus(newStatus)
			if err := s.repos.OperatorStatus.Create(ctx, status); err != nil {
				return nil, err
			}
			return status, nil
		}
		return nil, err
	}

	previousStatus := status.Status
	if previousStatus == newStatus {
		return status, nil // Idempotent
	}

	status.SetStatus(newStatus)
	if err := s.repos.OperatorStatus.Update(ctx, status); err != nil {
		return nil, err
	}

	// Grace period logic
	if previousStatus == domain.OperatorStatusAvailable && newStatus == domain.OperatorStatusOffline {
		s.createGracePeriods(ctx, operatorID)
	} else if previousStatus == domain.OperatorStatusOffline && newStatus == domain.OperatorStatusAvailable {
		s.repos.GracePeriodAssignments.DeleteByOperatorID(ctx, operatorID)
	}

	return status, nil
}

func (s *OperatorService) createGracePeriods(ctx context.Context, operatorID uuid.UUID) {
	operator, err := s.repos.Operators.GetByID(ctx, operatorID)
	if err != nil {
		s.logger.Warn("Failed to get operator for grace period creation",
			zap.String("operator_id", operatorID.String()),
			zap.Error(err))
		return
	}

	state := domain.ConversationStateAllocated
	conversations, err := s.repos.ConversationRefs.GetByOperatorID(ctx, operator.TenantID, operatorID, &state)
	if err != nil {
		s.logger.Warn("Failed to get conversations for grace period creation",
			zap.String("operator_id", operatorID.String()),
			zap.Error(err))
		return
	}

	expiresAt := time.Now().UTC().Add(GracePeriodDuration)
	for _, conv := range conversations {
		gpa := domain.NewGracePeriodAssignment(conv.ID, operatorID, expiresAt, domain.GracePeriodReasonOffline)
		if err := s.repos.GracePeriodAssignments.Create(ctx, gpa); err != nil {
			s.logger.Warn("Failed to create grace period for conversation",
				zap.String("conversation_id", conv.ID.String()),
				zap.Error(err))
		}
	}

	s.logger.Info("Grace periods created",
		zap.String("operator_id", operatorID.String()),
		zap.Int("count", len(conversations)))
}

// ==================== CRUD ====================

func (s *OperatorService) Create(ctx context.Context, tenantID uuid.UUID, role domain.OperatorRole) (*domain.Operator, error) {
	operator := domain.NewOperator(tenantID, role)
	if err := s.repos.Operators.Create(ctx, operator); err != nil {
		return nil, err
	}

	// Create initial status (OFFLINE by default)
	status := domain.NewOperatorStatus(operator.ID)
	if err := s.repos.OperatorStatus.Create(ctx, status); err != nil {
		s.logger.Warn("Failed to create initial operator status",
			zap.String("operator_id", operator.ID.String()),
			zap.Error(err))
	}

	return operator, nil
}

func (s *OperatorService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Operator, error) {
	return s.repos.Operators.GetByID(ctx, id)
}

func (s *OperatorService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Operator, error) {
	return s.repos.Operators.GetByTenantID(ctx, tenantID)
}

func (s *OperatorService) Update(ctx context.Context, id uuid.UUID, role domain.OperatorRole) (*domain.Operator, error) {
	operator, err := s.repos.Operators.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	operator.Role = role
	operator.UpdatedAt = time.Now().UTC()

	if err := s.repos.Operators.Update(ctx, operator); err != nil {
		return nil, err
	}
	return operator, nil
}

func (s *OperatorService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repos.Operators.Delete(ctx, id)
}
