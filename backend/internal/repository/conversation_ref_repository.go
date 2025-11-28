package repository

import (
	"context"
	"fmt"

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

// ListWithFilters returns conversations matching the given filters with cursor pagination
func (r *ConversationRefRepositoryImpl) ListWithFilters(ctx context.Context, filters ConversationFilters) ([]*domain.ConversationRef, error) {
	// Build dynamic query
	query := `
		SELECT 
			id, tenant_id, inbox_id, external_conversation_id,
			customer_phone_number, state, assigned_operator_id,
			last_message_at, message_count, priority_score,
			created_at, updated_at, resolved_at
		FROM conversation_refs
		WHERE tenant_id = $1
	`
	args := []interface{}{filters.TenantID}
	argIndex := 2

	// State filter
	if filters.State != nil {
		query += fmt.Sprintf(` AND state = $%d`, argIndex)
		args = append(args, string(*filters.State))
		argIndex++
	}

	// Inbox filter
	if filters.InboxID != nil {
		query += fmt.Sprintf(` AND inbox_id = $%d`, argIndex)
		args = append(args, *filters.InboxID)
		argIndex++
	}

	// Operator filter
	if filters.OperatorID != nil {
		query += fmt.Sprintf(` AND assigned_operator_id = $%d`, argIndex)
		args = append(args, *filters.OperatorID)
		argIndex++
	}

	// Allowed inboxes filter (for operators)
	if len(filters.AllowedInboxIDs) > 0 {
		query += fmt.Sprintf(` AND inbox_id = ANY($%d)`, argIndex)
		args = append(args, filters.AllowedInboxIDs)
		argIndex++
	}

	// Label filter (join)
	if filters.LabelID != nil {
		query += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM conversation_labels cl WHERE cl.conversation_id = id AND cl.label_id = $%d)`, argIndex)
		args = append(args, *filters.LabelID)
		argIndex++
	}

	// Cursor pagination
	if filters.HasCursor() {
		switch filters.SortOrder {
		case "oldest":
			query += fmt.Sprintf(` AND (last_message_at, id) > ($%d, $%d)`, argIndex, argIndex+1)
		case "priority":
			query += fmt.Sprintf(` AND (priority_score, last_message_at, id) < ($%d, $%d, $%d)`, argIndex, argIndex+1, argIndex+2)
		default: // newest
			query += fmt.Sprintf(` AND (last_message_at, id) < ($%d, $%d)`, argIndex, argIndex+1)
		}
		args = append(args, *filters.CursorTimestamp, *filters.CursorID)
		argIndex += 2
	}

	// Sorting
	switch filters.SortOrder {
	case "oldest":
		query += ` ORDER BY last_message_at ASC, id ASC`
	case "priority":
		query += ` ORDER BY priority_score DESC, last_message_at DESC, id DESC`
	default: // newest
		query += ` ORDER BY last_message_at DESC, id DESC`
	}

	// Limit
	query += fmt.Sprintf(` LIMIT $%d`, argIndex)
	args = append(args, filters.GetLimit())

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var conversations []*domain.ConversationRef
	for rows.Next() {
		var row ConversationRef
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.InboxID, &row.ExternalConversationID,
			&row.CustomerPhoneNumber, &row.State, &row.AssignedOperatorID,
			&row.LastMessageAt, &row.MessageCount, &row.PriorityScore,
			&row.CreatedAt, &row.UpdatedAt, &row.ResolvedAt,
		)
		if err != nil {
			return nil, mapError(err)
		}
		conversations = append(conversations, r.toDomain(row))
	}

	return conversations, nil
}

// GetByPhone returns conversations by customer phone number
func (r *ConversationRefRepositoryImpl) GetByPhone(ctx context.Context, tenantID uuid.UUID, phone string) ([]*domain.ConversationRef, error) {
	rows, err := r.q.SearchConversationsByPhone(ctx, SearchConversationsByPhoneParams{
		TenantID:            uuidToPgtype(tenantID),
		CustomerPhoneNumber: phone,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return r.toDomainSlice(rows), nil
}
