package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

type LabelRepositoryImpl struct {
	q *Queries
}

func NewLabelRepository(q *Queries) *LabelRepositoryImpl {
	return &LabelRepositoryImpl{q: q}
}

func (r *LabelRepositoryImpl) Create(ctx context.Context, label *domain.Label) error {
	return r.q.CreateLabel(ctx, CreateLabelParams{
		ID:        uuidToPgtype(label.ID),
		TenantID:  uuidToPgtype(label.TenantID),
		InboxID:   uuidToPgtype(label.InboxID),
		Name:      label.Name,
		Color:     stringPtrToPgtype(label.Color),
		CreatedBy: uuidPtrToPgtype(label.CreatedBy),
		CreatedAt: timeToPgtype(label.CreatedAt),
	})
}

func (r *LabelRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Label, error) {
	row, err := r.q.GetLabelByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *LabelRepositoryImpl) GetByInboxID(ctx context.Context, tenantID, inboxID uuid.UUID) ([]*domain.Label, error) {
	rows, err := r.q.GetLabelsByInboxID(ctx, GetLabelsByInboxIDParams{
		TenantID: uuidToPgtype(tenantID),
		InboxID:  uuidToPgtype(inboxID),
	})
	if err != nil {
		return nil, mapError(err)
	}

	labels := make([]*domain.Label, len(rows))
	for i, row := range rows {
		labels[i] = r.toDomain(row)
	}
	return labels, nil
}

func (r *LabelRepositoryImpl) GetByName(ctx context.Context, inboxID uuid.UUID, name string) (*domain.Label, error) {
	row, err := r.q.GetLabelByName(ctx, GetLabelByNameParams{
		InboxID: uuidToPgtype(inboxID),
		Name:    name,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *LabelRepositoryImpl) Update(ctx context.Context, label *domain.Label) error {
	return r.q.UpdateLabel(ctx, UpdateLabelParams{
		ID:    uuidToPgtype(label.ID),
		Name:  label.Name,
		Color: stringPtrToPgtype(label.Color),
	})
}

func (r *LabelRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteLabel(ctx, uuidToPgtype(id))
}

func (r *LabelRepositoryImpl) toDomain(row Label) *domain.Label {
	return &domain.Label{
		ID:        pgtypeToUUID(row.ID),
		TenantID:  pgtypeToUUID(row.TenantID),
		InboxID:   pgtypeToUUID(row.InboxID),
		Name:      row.Name,
		Color:     pgtypeToStringPtr(row.Color),
		CreatedBy: pgtypeToUUIDPtr(row.CreatedBy),
		CreatedAt: pgtypeToTime(row.CreatedAt),
	}
}
