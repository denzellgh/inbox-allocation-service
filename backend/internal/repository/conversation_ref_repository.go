package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConversationRefRepositoryImpl struct {
	q    *Queries
	pool *pgxpool.Pool
}

func NewConversationRefRepository(q *Queries, pool *pgxpool.Pool) *ConversationRefRepositoryImpl {
	return &ConversationRefRepositoryImpl{q: q, pool: pool}
}

func (r *ConversationRefRepositoryImpl) Create(ctx context.Context, conv *domain.ConversationRef) error {
	return r.q.CreateConversationRef(ctx, CreateConversationRefParams{
		ID:                     uuidToPgtype(conv.ID),
		TenantID:               uuidToPgtype(conv.TenantID),
		InboxID:                uuidToPgtype(conv.InboxID),
		ExternalConversationID: conv.ExternalConversationID,
		CustomerPhoneNumber:    conv.CustomerPhoneNumber,
		State:                  conversationStateToPgtype(conv.State),
		AssignedOperatorID:     uuidPtrToPgtype(conv.AssignedOperatorID),
		LastMessageAt:          timeToPgtype(conv.LastMessageAt),
		MessageCount:           conv.MessageCount,
		PriorityScore:          decimalToPgtype(conv.PriorityScore),
		CreatedAt:              timeToPgtype(conv.CreatedAt),
		UpdatedAt:              timeToPgtype(conv.UpdatedAt),
		ResolvedAt:             timePtrToPgtype(conv.ResolvedAt),
	})
}

func (r *ConversationRefRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.ConversationRef, error) {
	row, err := r.q.GetConversationRefByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *ConversationRefRepositoryImpl) GetByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*domain.ConversationRef, error) {
	row, err := r.q.GetConversationRefByExternalID(ctx, GetConversationRefByExternalIDParams{
		TenantID:               uuidToPgtype(tenantID),
		ExternalConversationID: externalID,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *ConversationRefRepositoryImpl) GetByFilter(ctx context.Context, filter domain.ConversationFilter) ([]*domain.ConversationRef, error) {
	// For now, implement basic filtering - can be extended with dynamic query building
	if filter.State != nil && *filter.State == domain.ConversationStateQueued {
		rows, err := r.q.GetQueuedConversationsByTenant(ctx, GetQueuedConversationsByTenantParams{
			TenantID: uuidToPgtype(filter.TenantID),
			Limit:    int32(filter.Limit),
		})
		if err != nil {
			return nil, mapError(err)
		}
		return r.toDomainSlice(rows), nil
	}

	if filter.InboxID != nil {
		rows, err := r.q.GetConversationsByInbox(ctx, GetConversationsByInboxParams{
			TenantID: uuidToPgtype(filter.TenantID),
			InboxID:  uuidToPgtype(*filter.InboxID),
			Limit:    int32(filter.Limit),
		})
		if err != nil {
			return nil, mapError(err)
		}
		return r.toDomainSlice(rows), nil
	}

	if filter.State != nil {
		rows, err := r.q.GetConversationsByTenantAndState(ctx, GetConversationsByTenantAndStateParams{
			TenantID: uuidToPgtype(filter.TenantID),
			State:    conversationStateToPgtype(*filter.State),
			Limit:    int32(filter.Limit),
		})
		if err != nil {
			return nil, mapError(err)
		}
		return r.toDomainSlice(rows), nil
	}

	// Default: return empty slice
	return []*domain.ConversationRef{}, nil
}

func (r *ConversationRefRepositoryImpl) SearchByPhone(ctx context.Context, tenantID uuid.UUID, phoneNumber string) ([]*domain.ConversationRef, error) {
	rows, err := r.q.SearchConversationsByPhone(ctx, SearchConversationsByPhoneParams{
		TenantID:            uuidToPgtype(tenantID),
		CustomerPhoneNumber: phoneNumber,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomainSlice(rows), nil
}

func (r *ConversationRefRepositoryImpl) Update(ctx context.Context, conv *domain.ConversationRef) error {
	return r.q.UpdateConversationRef(ctx, UpdateConversationRefParams{
		ID:                 uuidToPgtype(conv.ID),
		InboxID:            uuidToPgtype(conv.InboxID),
		State:              conversationStateToPgtype(conv.State),
		AssignedOperatorID: uuidPtrToPgtype(conv.AssignedOperatorID),
		LastMessageAt:      timeToPgtype(conv.LastMessageAt),
		MessageCount:       conv.MessageCount,
		PriorityScore:      decimalToPgtype(conv.PriorityScore),
		UpdatedAt:          timeToPgtype(conv.UpdatedAt),
		ResolvedAt:         timePtrToPgtype(conv.ResolvedAt),
	})
}

func (r *ConversationRefRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteConversationRef(ctx, uuidToPgtype(id))
}

// GetNextForAllocation - CRITICAL: Uses FOR UPDATE SKIP LOCKED
func (r *ConversationRefRepositoryImpl) GetNextForAllocation(ctx context.Context, tenantID uuid.UUID, inboxIDs []uuid.UUID, limit int) ([]*domain.ConversationRef, error) {
	// Convert []uuid.UUID to []pgtype.UUID
	pgtypeIDs := make([]pgtype.UUID, len(inboxIDs))
	for i, id := range inboxIDs {
		pgtypeIDs[i] = uuidToPgtype(id)
	}

	rows, err := r.q.GetNextConversationsForAllocation(ctx, GetNextConversationsForAllocationParams{
		TenantID: uuidToPgtype(tenantID),
		Column2:  pgtypeIDs,
		Limit:    int32(limit),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomainSlice(rows), nil
}

// LockForClaim - CRITICAL: Uses FOR UPDATE NOWAIT
func (r *ConversationRefRepositoryImpl) LockForClaim(ctx context.Context, id uuid.UUID) (*domain.ConversationRef, error) {
	row, err := r.q.LockConversationForClaim(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomain(row), nil
}

func (r *ConversationRefRepositoryImpl) GetByOperatorID(ctx context.Context, tenantID, operatorID uuid.UUID, state *domain.ConversationState) ([]*domain.ConversationRef, error) {
	if state != nil {
		rows, err := r.q.GetConversationsByOperatorAndState(ctx, GetConversationsByOperatorAndStateParams{
			TenantID:           uuidToPgtype(tenantID),
			AssignedOperatorID: uuidToPgtype(operatorID),
			State:              conversationStateToPgtype(*state),
		})
		if err != nil {
			return nil, mapError(err)
		}
		return r.toDomainSlice(rows), nil
	}

	rows, err := r.q.GetConversationsByOperatorID(ctx, GetConversationsByOperatorIDParams{
		TenantID:           uuidToPgtype(tenantID),
		AssignedOperatorID: uuidToPgtype(operatorID),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomainSlice(rows), nil
}

func (r *ConversationRefRepositoryImpl) toDomain(row ConversationRef) *domain.ConversationRef {
	return &domain.ConversationRef{
		ID:                     pgtypeToUUID(row.ID),
		TenantID:               pgtypeToUUID(row.TenantID),
		InboxID:                pgtypeToUUID(row.InboxID),
		ExternalConversationID: row.ExternalConversationID,
		CustomerPhoneNumber:    row.CustomerPhoneNumber,
		State:                  pgtypeToConversationState(row.State),
		AssignedOperatorID:     pgtypeToUUIDPtr(row.AssignedOperatorID),
		LastMessageAt:          pgtypeToTime(row.LastMessageAt),
		MessageCount:           row.MessageCount,
		PriorityScore:          pgtypeToDecimal(row.PriorityScore),
		CreatedAt:              pgtypeToTime(row.CreatedAt),
		UpdatedAt:              pgtypeToTime(row.UpdatedAt),
		ResolvedAt:             pgtypeToTimePtr(row.ResolvedAt),
	}
}

func (r *ConversationRefRepositoryImpl) toDomainSlice(rows []ConversationRef) []*domain.ConversationRef {
	conversations := make([]*domain.ConversationRef, len(rows))
	for i, row := range rows {
		conversations[i] = r.toDomain(row)
	}
	return conversations
}
