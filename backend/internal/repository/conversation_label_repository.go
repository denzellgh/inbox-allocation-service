package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

type ConversationLabelRepositoryImpl struct {
	q *Queries
}

func NewConversationLabelRepository(q *Queries) *ConversationLabelRepositoryImpl {
	return &ConversationLabelRepositoryImpl{q: q}
}

func (r *ConversationLabelRepositoryImpl) Create(ctx context.Context, cl *domain.ConversationLabel) error {
	return r.q.CreateConversationLabel(ctx, CreateConversationLabelParams{
		ID:             uuidToPgtype(cl.ID),
		ConversationID: uuidToPgtype(cl.ConversationID),
		LabelID:        uuidToPgtype(cl.LabelID),
		CreatedAt:      timeToPgtype(cl.CreatedAt),
	})
}

func (r *ConversationLabelRepositoryImpl) GetByConversationID(ctx context.Context, conversationID uuid.UUID) ([]*domain.ConversationLabel, error) {
	rows, err := r.q.GetConversationLabelsByConversationID(ctx, uuidToPgtype(conversationID))
	if err != nil {
		return nil, mapError(err)
	}

	labels := make([]*domain.ConversationLabel, len(rows))
	for i, row := range rows {
		labels[i] = r.toDomain(row)
	}
	return labels, nil
}

func (r *ConversationLabelRepositoryImpl) GetByLabelID(ctx context.Context, labelID uuid.UUID) ([]*domain.ConversationLabel, error) {
	rows, err := r.q.GetConversationLabelsByLabelID(ctx, uuidToPgtype(labelID))
	if err != nil {
		return nil, mapError(err)
	}

	labels := make([]*domain.ConversationLabel, len(rows))
	for i, row := range rows {
		labels[i] = r.toDomain(row)
	}
	return labels, nil
}

func (r *ConversationLabelRepositoryImpl) Delete(ctx context.Context, conversationID, labelID uuid.UUID) error {
	return r.q.DeleteConversationLabel(ctx, DeleteConversationLabelParams{
		ConversationID: uuidToPgtype(conversationID),
		LabelID:        uuidToPgtype(labelID),
	})
}

func (r *ConversationLabelRepositoryImpl) DeleteAllForConversation(ctx context.Context, conversationID uuid.UUID) error {
	return r.q.DeleteAllConversationLabels(ctx, uuidToPgtype(conversationID))
}

func (r *ConversationLabelRepositoryImpl) Exists(ctx context.Context, conversationID, labelID uuid.UUID) (bool, error) {
	exists, err := r.q.CheckConversationLabelExists(ctx, CheckConversationLabelExistsParams{
		ConversationID: uuidToPgtype(conversationID),
		LabelID:        uuidToPgtype(labelID),
	})
	if err != nil {
		return false, mapError(err)
	}
	return exists, nil
}

func (r *ConversationLabelRepositoryImpl) toDomain(row ConversationLabel) *domain.ConversationLabel {
	return &domain.ConversationLabel{
		ID:             pgtypeToUUID(row.ID),
		ConversationID: pgtypeToUUID(row.ConversationID),
		LabelID:        pgtypeToUUID(row.LabelID),
		CreatedAt:      pgtypeToTime(row.CreatedAt),
	}
}
