package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/repository"
	"go.uber.org/zap"
)

type SubscriptionService struct {
	repos  *repository.RepositoryContainer
	logger *zap.Logger
}

func NewSubscriptionService(repos *repository.RepositoryContainer, logger *zap.Logger) *SubscriptionService {
	return &SubscriptionService{repos: repos, logger: logger}
}

func (s *SubscriptionService) Subscribe(ctx context.Context, operatorID, inboxID uuid.UUID) (*domain.OperatorInboxSubscription, error) {
	isSubscribed, err := s.repos.Subscriptions.IsSubscribed(ctx, operatorID, inboxID)
	if err != nil {
		return nil, err
	}
	if isSubscribed {
		// Idempotent: return existing subscription
		return s.repos.Subscriptions.GetByOperatorAndInbox(ctx, operatorID, inboxID)
	}

	if _, err := s.repos.Operators.GetByID(ctx, operatorID); err != nil {
		return nil, err
	}
	if _, err := s.repos.Inboxes.GetByID(ctx, inboxID); err != nil {
		return nil, err
	}

	sub := domain.NewOperatorInboxSubscription(operatorID, inboxID)
	if err := s.repos.Subscriptions.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) Unsubscribe(ctx context.Context, operatorID, inboxID uuid.UUID) error {
	return s.repos.Subscriptions.DeleteByOperatorAndInbox(ctx, operatorID, inboxID)
}

func (s *SubscriptionService) GetOperatorsByInbox(ctx context.Context, inboxID uuid.UUID) ([]*domain.OperatorInboxSubscription, error) {
	return s.repos.Subscriptions.GetByInboxID(ctx, inboxID)
}

func (s *SubscriptionService) GetInboxesByOperator(ctx context.Context, operatorID uuid.UUID) ([]*domain.OperatorInboxSubscription, error) {
	return s.repos.Subscriptions.GetByOperatorID(ctx, operatorID)
}

func (s *SubscriptionService) IsSubscribed(ctx context.Context, operatorID, inboxID uuid.UUID) (bool, error) {
	return s.repos.Subscriptions.IsSubscribed(ctx, operatorID, inboxID)
}
