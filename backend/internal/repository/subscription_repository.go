package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

type SubscriptionRepositoryImpl struct {
	q *Queries
}

func NewSubscriptionRepository(q *Queries) *SubscriptionRepositoryImpl {
	return &SubscriptionRepositoryImpl{q: q}
}

func (r *SubscriptionRepositoryImpl) Create(ctx context.Context, sub *domain.OperatorInboxSubscription) error {
	return r.q.CreateSubscription(ctx, CreateSubscriptionParams{
		ID:         uuidToPgtype(sub.ID),
		OperatorID: uuidToPgtype(sub.OperatorID),
		InboxID:    uuidToPgtype(sub.InboxID),
		CreatedAt:  timeToPgtype(sub.CreatedAt),
	})
}

func (r *SubscriptionRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.OperatorInboxSubscription, error) {
	row, err := r.q.GetSubscriptionByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *SubscriptionRepositoryImpl) GetByOperatorID(ctx context.Context, operatorID uuid.UUID) ([]*domain.OperatorInboxSubscription, error) {
	rows, err := r.q.GetSubscriptionsByOperatorID(ctx, uuidToPgtype(operatorID))
	if err != nil {
		return nil, mapError(err)
	}

	subs := make([]*domain.OperatorInboxSubscription, len(rows))
	for i, row := range rows {
		subs[i] = r.toDomain(row)
	}
	return subs, nil
}

func (r *SubscriptionRepositoryImpl) GetByInboxID(ctx context.Context, inboxID uuid.UUID) ([]*domain.OperatorInboxSubscription, error) {
	rows, err := r.q.GetSubscriptionsByInboxID(ctx, uuidToPgtype(inboxID))
	if err != nil {
		return nil, mapError(err)
	}

	subs := make([]*domain.OperatorInboxSubscription, len(rows))
	for i, row := range rows {
		subs[i] = r.toDomain(row)
	}
	return subs, nil
}

func (r *SubscriptionRepositoryImpl) GetByOperatorAndInbox(ctx context.Context, operatorID, inboxID uuid.UUID) (*domain.OperatorInboxSubscription, error) {
	row, err := r.q.GetSubscriptionByOperatorAndInbox(ctx, GetSubscriptionByOperatorAndInboxParams{
		OperatorID: uuidToPgtype(operatorID),
		InboxID:    uuidToPgtype(inboxID),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *SubscriptionRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteSubscription(ctx, uuidToPgtype(id))
}

func (r *SubscriptionRepositoryImpl) DeleteByOperatorAndInbox(ctx context.Context, operatorID, inboxID uuid.UUID) error {
	return r.q.DeleteSubscriptionByOperatorAndInbox(ctx, DeleteSubscriptionByOperatorAndInboxParams{
		OperatorID: uuidToPgtype(operatorID),
		InboxID:    uuidToPgtype(inboxID),
	})
}

func (r *SubscriptionRepositoryImpl) GetSubscribedInboxIDs(ctx context.Context, operatorID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.q.GetSubscribedInboxIDs(ctx, uuidToPgtype(operatorID))
	if err != nil {
		return nil, mapError(err)
	}

	ids := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		ids[i] = pgtypeToUUID(row)
	}
	return ids, nil
}

func (r *SubscriptionRepositoryImpl) IsSubscribed(ctx context.Context, operatorID, inboxID uuid.UUID) (bool, error) {
	exists, err := r.q.CheckSubscriptionExists(ctx, CheckSubscriptionExistsParams{
		OperatorID: uuidToPgtype(operatorID),
		InboxID:    uuidToPgtype(inboxID),
	})
	if err != nil {
		return false, mapError(err)
	}
	return exists, nil
}

func (r *SubscriptionRepositoryImpl) toDomain(row OperatorInboxSubscription) *domain.OperatorInboxSubscription {
	return &domain.OperatorInboxSubscription{
		ID:         pgtypeToUUID(row.ID),
		OperatorID: pgtypeToUUID(row.OperatorID),
		InboxID:    pgtypeToUUID(row.InboxID),
		CreatedAt:  pgtypeToTime(row.CreatedAt),
	}
}
