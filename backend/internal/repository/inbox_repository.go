package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

type InboxRepositoryImpl struct {
	q *Queries
}

func NewInboxRepository(q *Queries) *InboxRepositoryImpl {
	return &InboxRepositoryImpl{q: q}
}

func (r *InboxRepositoryImpl) Create(ctx context.Context, inbox *domain.Inbox) error {
	return r.q.CreateInbox(ctx, CreateInboxParams{
		ID:          uuidToPgtype(inbox.ID),
		TenantID:    uuidToPgtype(inbox.TenantID),
		PhoneNumber: inbox.PhoneNumber,
		DisplayName: inbox.DisplayName,
		CreatedAt:   timeToPgtype(inbox.CreatedAt),
		UpdatedAt:   timeToPgtype(inbox.UpdatedAt),
	})
}

func (r *InboxRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Inbox, error) {
	row, err := r.q.GetInboxByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *InboxRepositoryImpl) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.Inbox, error) {
	rows, err := r.q.GetInboxesByTenantID(ctx, uuidToPgtype(tenantID))
	if err != nil {
		return nil, mapError(err)
	}

	inboxes := make([]*domain.Inbox, len(rows))
	for i, row := range rows {
		inboxes[i] = r.toDomain(row)
	}
	return inboxes, nil
}

func (r *InboxRepositoryImpl) GetByPhoneNumber(ctx context.Context, tenantID uuid.UUID, phoneNumber string) (*domain.Inbox, error) {
	row, err := r.q.GetInboxByPhoneNumber(ctx, GetInboxByPhoneNumberParams{
		TenantID:    uuidToPgtype(tenantID),
		PhoneNumber: phoneNumber,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *InboxRepositoryImpl) Update(ctx context.Context, inbox *domain.Inbox) error {
	return r.q.UpdateInbox(ctx, UpdateInboxParams{
		ID:          uuidToPgtype(inbox.ID),
		PhoneNumber: inbox.PhoneNumber,
		DisplayName: inbox.DisplayName,
		UpdatedAt:   timeToPgtype(inbox.UpdatedAt),
	})
}

func (r *InboxRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteInbox(ctx, uuidToPgtype(id))
}

func (r *InboxRepositoryImpl) toDomain(row Inbox) *domain.Inbox {
	return &domain.Inbox{
		ID:          pgtypeToUUID(row.ID),
		TenantID:    pgtypeToUUID(row.TenantID),
		PhoneNumber: row.PhoneNumber,
		DisplayName: row.DisplayName,
		CreatedAt:   pgtypeToTime(row.CreatedAt),
		UpdatedAt:   pgtypeToTime(row.UpdatedAt),
	}
}
