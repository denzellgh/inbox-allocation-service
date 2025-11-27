package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/repository"
	"go.uber.org/zap"
)

type InboxService struct {
	repos  *repository.RepositoryContainer
	logger *zap.Logger
}

func NewInboxService(repos *repository.RepositoryContainer, logger *zap.Logger) *InboxService {
	return &InboxService{repos: repos, logger: logger}
}

func (s *InboxService) Create(ctx context.Context, tenantID uuid.UUID, phoneNumber, displayName string) (*domain.Inbox, error) {
	existing, err := s.repos.Inboxes.GetByPhoneNumber(ctx, tenantID, phoneNumber)
	if err == nil && existing != nil {
		return nil, domain.ErrAlreadyExists
	}

	inbox := domain.NewInbox(tenantID, phoneNumber, displayName)
	if err := s.repos.Inboxes.Create(ctx, inbox); err != nil {
		return nil, err
	}
	return inbox, nil
}

func (s *InboxService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Inbox, error) {
	return s.repos.Inboxes.GetByID(ctx, id)
}

func (s *InboxService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Inbox, error) {
	return s.repos.Inboxes.GetByTenantID(ctx, tenantID)
}

func (s *InboxService) ListForOperator(ctx context.Context, tenantID, operatorID uuid.UUID) ([]*domain.Inbox, error) {
	inboxIDs, err := s.repos.Subscriptions.GetSubscribedInboxIDs(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	if len(inboxIDs) == 0 {
		return []*domain.Inbox{}, nil
	}

	var inboxes []*domain.Inbox
	for _, id := range inboxIDs {
		inbox, err := s.repos.Inboxes.GetByID(ctx, id)
		if err != nil {
			if err == domain.ErrNotFound {
				continue // Skip deleted inboxes
			}
			return nil, err
		}
		if inbox.TenantID == tenantID {
			inboxes = append(inboxes, inbox)
		}
	}
	return inboxes, nil
}

func (s *InboxService) Update(ctx context.Context, id uuid.UUID, phoneNumber, displayName *string) (*domain.Inbox, error) {
	inbox, err := s.repos.Inboxes.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if phoneNumber != nil {
		existing, err := s.repos.Inboxes.GetByPhoneNumber(ctx, inbox.TenantID, *phoneNumber)
		if err == nil && existing != nil && existing.ID != id {
			return nil, domain.ErrAlreadyExists
		}
		inbox.PhoneNumber = *phoneNumber
	}

	if displayName != nil {
		inbox.DisplayName = *displayName
	}

	inbox.UpdatedAt = time.Now().UTC()
	if err := s.repos.Inboxes.Update(ctx, inbox); err != nil {
		return nil, err
	}
	return inbox, nil
}

func (s *InboxService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repos.Inboxes.Delete(ctx, id)
}
