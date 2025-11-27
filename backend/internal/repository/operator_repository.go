package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

type OperatorRepositoryImpl struct {
	q *Queries
}

func NewOperatorRepository(q *Queries) *OperatorRepositoryImpl {
	return &OperatorRepositoryImpl{q: q}
}

func (r *OperatorRepositoryImpl) Create(ctx context.Context, operator *domain.Operator) error {
	return r.q.CreateOperator(ctx, CreateOperatorParams{
		ID:        uuidToPgtype(operator.ID),
		TenantID:  uuidToPgtype(operator.TenantID),
		Role:      operatorRoleToPgtype(operator.Role),
		CreatedAt: timeToPgtype(operator.CreatedAt),
		UpdatedAt: timeToPgtype(operator.UpdatedAt),
	})
}

func (r *OperatorRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Operator, error) {
	row, err := r.q.GetOperatorByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *OperatorRepositoryImpl) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.Operator, error) {
	rows, err := r.q.GetOperatorsByTenantID(ctx, uuidToPgtype(tenantID))
	if err != nil {
		return nil, mapError(err)
	}

	operators := make([]*domain.Operator, len(rows))
	for i, row := range rows {
		operators[i] = r.toDomain(row)
	}
	return operators, nil
}

func (r *OperatorRepositoryImpl) GetByTenantAndRole(ctx context.Context, tenantID uuid.UUID, role domain.OperatorRole) ([]*domain.Operator, error) {
	rows, err := r.q.GetOperatorsByTenantAndRole(ctx, GetOperatorsByTenantAndRoleParams{
		TenantID: uuidToPgtype(tenantID),
		Role:     operatorRoleToPgtype(role),
	})
	if err != nil {
		return nil, mapError(err)
	}

	operators := make([]*domain.Operator, len(rows))
	for i, row := range rows {
		operators[i] = r.toDomain(row)
	}
	return operators, nil
}

func (r *OperatorRepositoryImpl) Update(ctx context.Context, operator *domain.Operator) error {
	return r.q.UpdateOperator(ctx, UpdateOperatorParams{
		ID:        uuidToPgtype(operator.ID),
		Role:      operatorRoleToPgtype(operator.Role),
		UpdatedAt: timeToPgtype(operator.UpdatedAt),
	})
}

func (r *OperatorRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteOperator(ctx, uuidToPgtype(id))
}

func (r *OperatorRepositoryImpl) toDomain(row Operator) *domain.Operator {
	return &domain.Operator{
		ID:        pgtypeToUUID(row.ID),
		TenantID:  pgtypeToUUID(row.TenantID),
		Role:      pgtypeToOperatorRole(row.Role),
		CreatedAt: pgtypeToTime(row.CreatedAt),
		UpdatedAt: pgtypeToTime(row.UpdatedAt),
	}
}
