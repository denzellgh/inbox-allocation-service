package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GracePeriodRepositoryImpl struct {
	q    *Queries
	pool *pgxpool.Pool
}

func NewGracePeriodRepository(q *Queries, pool *pgxpool.Pool) *GracePeriodRepositoryImpl {
	return &GracePeriodRepositoryImpl{q: q, pool: pool}
}

func (r *GracePeriodRepositoryImpl) Create(ctx context.Context, gpa *domain.GracePeriodAssignment) error {
	return r.q.CreateGracePeriodAssignment(ctx, CreateGracePeriodAssignmentParams{
		ID:             uuidToPgtype(gpa.ID),
		ConversationID: uuidToPgtype(gpa.ConversationID),
		OperatorID:     uuidToPgtype(gpa.OperatorID),
		ExpiresAt:      timeToPgtype(gpa.ExpiresAt),
		Reason:         gracePeriodReasonToPgtype(gpa.Reason),
		CreatedAt:      timeToPgtype(gpa.CreatedAt),
	})
}

func (r *GracePeriodRepositoryImpl) GetByConversationID(ctx context.Context, conversationID uuid.UUID) (*domain.GracePeriodAssignment, error) {
	row, err := r.q.GetGracePeriodByConversationID(ctx, uuidToPgtype(conversationID))
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *GracePeriodRepositoryImpl) GetByOperatorID(ctx context.Context, operatorID uuid.UUID) ([]*domain.GracePeriodAssignment, error) {
	rows, err := r.q.GetGracePeriodsByOperatorID(ctx, uuidToPgtype(operatorID))
	if err != nil {
		return nil, mapError(err)
	}

	assignments := make([]*domain.GracePeriodAssignment, len(rows))
	for i, row := range rows {
		assignments[i] = r.toDomain(row)
	}
	return assignments, nil
}

func (r *GracePeriodRepositoryImpl) GetExpired(ctx context.Context, limit int) ([]*domain.GracePeriodAssignment, error) {
	rows, err := r.q.GetExpiredGracePeriods(ctx, int32(limit))
	if err != nil {
		return nil, mapError(err)
	}

	assignments := make([]*domain.GracePeriodAssignment, len(rows))
	for i, row := range rows {
		assignments[i] = r.toDomain(row)
	}
	return assignments, nil
}

func (r *GracePeriodRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteGracePeriodAssignment(ctx, uuidToPgtype(id))
}

func (r *GracePeriodRepositoryImpl) DeleteByOperatorID(ctx context.Context, operatorID uuid.UUID) error {
	return r.q.DeleteGracePeriodsByOperatorID(ctx, uuidToPgtype(operatorID))
}

func (r *GracePeriodRepositoryImpl) DeleteByConversationID(ctx context.Context, conversationID uuid.UUID) error {
	return r.q.DeleteGracePeriodByConversationID(ctx, uuidToPgtype(conversationID))
}

// GetAndLockExpired uses FOR UPDATE SKIP LOCKED for worker processing
func (r *GracePeriodRepositoryImpl) GetAndLockExpired(ctx context.Context, limit int) ([]*domain.GracePeriodAssignment, error) {
	rows, err := r.q.GetAndLockExpiredGracePeriods(ctx, int32(limit))
	if err != nil {
		return nil, mapError(err)
	}

	assignments := make([]*domain.GracePeriodAssignment, len(rows))
	for i, row := range rows {
		assignments[i] = r.toDomain(row)
	}
	return assignments, nil
}

func (r *GracePeriodRepositoryImpl) toDomain(row GracePeriodAssignment) *domain.GracePeriodAssignment {
	return &domain.GracePeriodAssignment{
		ID:             pgtypeToUUID(row.ID),
		ConversationID: pgtypeToUUID(row.ConversationID),
		OperatorID:     pgtypeToUUID(row.OperatorID),
		ExpiresAt:      pgtypeToTime(row.ExpiresAt),
		Reason:         pgtypeToGracePeriodReason(row.Reason),
		CreatedAt:      pgtypeToTime(row.CreatedAt),
	}
}
