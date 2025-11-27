package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

type OperatorStatusRepositoryImpl struct {
	q *Queries
}

func NewOperatorStatusRepository(q *Queries) *OperatorStatusRepositoryImpl {
	return &OperatorStatusRepositoryImpl{q: q}
}

func (r *OperatorStatusRepositoryImpl) Create(ctx context.Context, status *domain.OperatorStatus) error {
	return r.q.CreateOperatorStatus(ctx, CreateOperatorStatusParams{
		ID:                 uuidToPgtype(status.ID),
		OperatorID:         uuidToPgtype(status.OperatorID),
		Status:             operatorStatusTypeToPgtype(status.Status),
		LastStatusChangeAt: timeToPgtype(status.LastStatusChangeAt),
	})
}

func (r *OperatorStatusRepositoryImpl) GetByOperatorID(ctx context.Context, operatorID uuid.UUID) (*domain.OperatorStatus, error) {
	row, err := r.q.GetOperatorStatusByOperatorID(ctx, uuidToPgtype(operatorID))
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *OperatorStatusRepositoryImpl) Update(ctx context.Context, status *domain.OperatorStatus) error {
	return r.q.UpdateOperatorStatus(ctx, UpdateOperatorStatusParams{
		OperatorID:         uuidToPgtype(status.OperatorID),
		Status:             operatorStatusTypeToPgtype(status.Status),
		LastStatusChangeAt: timeToPgtype(status.LastStatusChangeAt),
	})
}

func (r *OperatorStatusRepositoryImpl) GetAvailableOperators(ctx context.Context, tenantID uuid.UUID) ([]*domain.OperatorStatus, error) {
	rows, err := r.q.GetAvailableOperators(ctx, uuidToPgtype(tenantID))
	if err != nil {
		return nil, mapError(err)
	}

	statuses := make([]*domain.OperatorStatus, len(rows))
	for i, row := range rows {
		statuses[i] = r.toDomain(row)
	}
	return statuses, nil
}

func (r *OperatorStatusRepositoryImpl) toDomain(row OperatorStatus) *domain.OperatorStatus {
	return &domain.OperatorStatus{
		ID:                 pgtypeToUUID(row.ID),
		OperatorID:         pgtypeToUUID(row.OperatorID),
		Status:             pgtypeToOperatorStatusType(row.Status),
		LastStatusChangeAt: pgtypeToTime(row.LastStatusChangeAt),
	}
}
