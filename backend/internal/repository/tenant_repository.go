package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

type TenantRepositoryImpl struct {
	q *Queries
}

func NewTenantRepository(q *Queries) *TenantRepositoryImpl {
	return &TenantRepositoryImpl{q: q}
}

func (r *TenantRepositoryImpl) Create(ctx context.Context, t *domain.Tenant) error {
	return r.q.CreateTenant(ctx, CreateTenantParams{
		ID:                  uuidToPgtype(t.ID),
		Name:                t.Name,
		PriorityWeightAlpha: decimalToPgtype(t.PriorityWeightAlpha),
		PriorityWeightBeta:  decimalToPgtype(t.PriorityWeightBeta),
		CreatedAt:           timeToPgtype(t.CreatedAt),
		UpdatedAt:           timeToPgtype(t.UpdatedAt),
		UpdatedBy:           uuidPtrToPgtype(t.UpdatedBy),
	})
}

func (r *TenantRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	row, err := r.q.GetTenantByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *TenantRepositoryImpl) GetByName(ctx context.Context, name string) (*domain.Tenant, error) {
	row, err := r.q.GetTenantByName(ctx, name)
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *TenantRepositoryImpl) Update(ctx context.Context, t *domain.Tenant) error {
	return r.q.UpdateTenant(ctx, UpdateTenantParams{
		ID:                  uuidToPgtype(t.ID),
		Name:                t.Name,
		PriorityWeightAlpha: decimalToPgtype(t.PriorityWeightAlpha),
		PriorityWeightBeta:  decimalToPgtype(t.PriorityWeightBeta),
		UpdatedAt:           timeToPgtype(t.UpdatedAt),
		UpdatedBy:           uuidPtrToPgtype(t.UpdatedBy),
	})
}

func (r *TenantRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteTenant(ctx, uuidToPgtype(id))
}

func (r *TenantRepositoryImpl) toDomain(row Tenant) *domain.Tenant {
	return &domain.Tenant{
		ID:                  pgtypeToUUID(row.ID),
		Name:                row.Name,
		PriorityWeightAlpha: pgtypeToDecimal(row.PriorityWeightAlpha),
		PriorityWeightBeta:  pgtypeToDecimal(row.PriorityWeightBeta),
		CreatedAt:           pgtypeToTime(row.CreatedAt),
		UpdatedAt:           pgtypeToTime(row.UpdatedAt),
		UpdatedBy:           pgtypeToUUIDPtr(row.UpdatedBy),
	}
}
