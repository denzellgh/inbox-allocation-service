package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

type IdempotencyRepositoryImpl struct {
	q *Queries
}

func NewIdempotencyRepository(q *Queries) *IdempotencyRepositoryImpl {
	return &IdempotencyRepositoryImpl{q: q}
}

func (r *IdempotencyRepositoryImpl) Create(ctx context.Context, ik *domain.IdempotencyKey) error {
	return r.q.CreateIdempotencyKey(ctx, CreateIdempotencyKeyParams{
		ID:             uuidToPgtype(ik.ID),
		Key:            ik.Key,
		TenantID:       uuidToPgtype(ik.TenantID),
		Endpoint:       ik.Endpoint,
		Method:         ik.Method,
		RequestHash:    stringPtrToPgtype(ik.RequestHash),
		ResponseStatus: int32(ik.ResponseStatus),
		ResponseBody:   ik.ResponseBody,
		CreatedAt:      timeToPgtype(ik.CreatedAt),
		ExpiresAt:      timeToPgtype(ik.ExpiresAt),
	})
}

func (r *IdempotencyRepositoryImpl) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*domain.IdempotencyKey, error) {
	row, err := r.q.GetIdempotencyKey(ctx, GetIdempotencyKeyParams{
		TenantID: uuidToPgtype(tenantID),
		Key:      key,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *IdempotencyRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteIdempotencyKey(ctx, uuidToPgtype(id))
}

func (r *IdempotencyRepositoryImpl) DeleteExpired(ctx context.Context) (int64, error) {
	return r.q.DeleteExpiredIdempotencyKeys(ctx)
}

func (r *IdempotencyRepositoryImpl) GetExpiredForCleanup(ctx context.Context, limit int) ([]*domain.IdempotencyKey, error) {
	rows, err := r.q.GetExpiredIdempotencyKeysForCleanup(ctx, int32(limit))
	if err != nil {
		return nil, mapError(err)
	}

	keys := make([]*domain.IdempotencyKey, len(rows))
	for i, row := range rows {
		keys[i] = r.toDomain(row)
	}
	return keys, nil
}

func (r *IdempotencyRepositoryImpl) toDomain(row IdempotencyKey) *domain.IdempotencyKey {
	return &domain.IdempotencyKey{
		ID:             pgtypeToUUID(row.ID),
		Key:            row.Key,
		TenantID:       pgtypeToUUID(row.TenantID),
		Endpoint:       row.Endpoint,
		Method:         row.Method,
		RequestHash:    pgtypeToStringPtr(row.RequestHash),
		ResponseStatus: int(row.ResponseStatus),
		ResponseBody:   row.ResponseBody,
		CreatedAt:      pgtypeToTime(row.CreatedAt),
		ExpiresAt:      pgtypeToTime(row.ExpiresAt),
	}
}
