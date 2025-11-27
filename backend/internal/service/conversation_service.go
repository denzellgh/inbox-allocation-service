package service

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/api/dto"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type ConversationService struct {
	repos  *repository.RepositoryContainer
	logger *zap.Logger
}

func NewConversationService(repos *repository.RepositoryContainer, logger *zap.Logger) *ConversationService {
	return &ConversationService{repos: repos, logger: logger}
}

// ==================== List Conversations ====================

type ListConversationsParams struct {
	TenantID   uuid.UUID
	OperatorID uuid.UUID
	Role       domain.OperatorRole

	// Filters
	State            *domain.ConversationState
	InboxID          *uuid.UUID
	OperatorFilterID *uuid.UUID
	LabelID          *uuid.UUID

	// Sorting
	Sort string

	// Pagination
	Cursor  *dto.Cursor
	PerPage int
}

func (s *ConversationService) List(ctx context.Context, params ListConversationsParams) ([]*domain.ConversationRef, error) {
	// Get allowed inbox IDs based on role
	var allowedInboxIDs []uuid.UUID

	if params.Role == domain.OperatorRoleOperator {
		// Operators can only see conversations in their subscribed inboxes
		ids, err := s.repos.Subscriptions.GetSubscribedInboxIDs(ctx, params.OperatorID)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return []*domain.ConversationRef{}, nil
		}
		allowedInboxIDs = ids
	}

	// Build query filters
	filters := repository.ConversationFilters{
		TenantID:        params.TenantID,
		State:           params.State,
		InboxID:         params.InboxID,
		OperatorID:      params.OperatorFilterID,
		LabelID:         params.LabelID,
		AllowedInboxIDs: allowedInboxIDs,
		Limit:           params.PerPage,
	}

	// Apply cursor for pagination
	if params.Cursor != nil {
		filters.CursorTimestamp = &params.Cursor.Timestamp
		filters.CursorID = &params.Cursor.ID
	}

	// Set sort order
	filters.SortOrder = params.Sort

	// Execute query
	conversations, err := s.repos.ConversationRefs.ListWithFilters(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to list conversations",
			zap.String("tenant_id", params.TenantID.String()),
			zap.Error(err))
		return nil, err
	}

	return conversations, nil
}

// ==================== Get Single Conversation ====================

func (s *ConversationService) GetByID(ctx context.Context, tenantID, conversationID uuid.UUID) (*domain.ConversationRef, error) {
	conv, err := s.repos.ConversationRefs.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Verify tenant
	if conv.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}

	return conv, nil
}

// CanAccess checks if operator can access the conversation
func (s *ConversationService) CanAccess(ctx context.Context, operatorID uuid.UUID, role domain.OperatorRole, conv *domain.ConversationRef) bool {
	// Managers and Admins can access all conversations in tenant
	if role == domain.OperatorRoleManager || role == domain.OperatorRoleAdmin {
		return true
	}

	// Operators can only access conversations in subscribed inboxes
	isSubscribed, err := s.repos.Subscriptions.IsSubscribed(ctx, operatorID, conv.InboxID)
	if err != nil {
		return false
	}

	return isSubscribed
}

// ==================== Search by Phone ====================

func (s *ConversationService) SearchByPhone(ctx context.Context, tenantID uuid.UUID, phone string, operatorID uuid.UUID, role domain.OperatorRole) ([]*domain.ConversationRef, error) {
	// Get conversations by phone
	conversations, err := s.repos.ConversationRefs.GetByPhone(ctx, tenantID, phone)
	if err != nil {
		return nil, err
	}

	// If not admin/manager, filter by subscribed inboxes
	if role == domain.OperatorRoleOperator {
		inboxIDs, err := s.repos.Subscriptions.GetSubscribedInboxIDs(ctx, operatorID)
		if err != nil {
			return nil, err
		}

		inboxSet := make(map[uuid.UUID]bool)
		for _, id := range inboxIDs {
			inboxSet[id] = true
		}

		filtered := make([]*domain.ConversationRef, 0)
		for _, conv := range conversations {
			if inboxSet[conv.InboxID] {
				filtered = append(filtered, conv)
			}
		}
		conversations = filtered
	}

	// Limit results
	if len(conversations) > dto.MaxConversationsPerQuery {
		conversations = conversations[:dto.MaxConversationsPerQuery]
	}

	return conversations, nil
}

// ==================== Priority Calculation ====================

// CalculatePriority computes the priority score for a conversation
// Formula: priority_score = (alpha × normalized_message_count) + (beta × normalized_delay)
func (s *ConversationService) CalculatePriority(ctx context.Context, tenantID uuid.UUID, conv *domain.ConversationRef) (decimal.Decimal, error) {
	// Get tenant weights
	tenant, err := s.repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		// Use default weights if tenant not found
		return s.calculatePriorityWithWeights(conv, decimal.NewFromFloat(0.5), decimal.NewFromFloat(0.5)), nil
	}

	return s.calculatePriorityWithWeights(conv, tenant.PriorityWeightAlpha, tenant.PriorityWeightBeta), nil
}

func (s *ConversationService) calculatePriorityWithWeights(conv *domain.ConversationRef, alpha, beta decimal.Decimal) decimal.Decimal {
	// Normalize message count: min(log10(message_count + 1) / 3, 1.0)
	normalizedMessageCount := math.Min(math.Log10(float64(conv.MessageCount+1))/3.0, 1.0)

	// Normalize delay: min(hours_since_last_message / 24, 1.0)
	hoursSinceLastMessage := time.Since(conv.LastMessageAt).Hours()
	normalizedDelay := math.Min(hoursSinceLastMessage/24.0, 1.0)

	// Calculate priority: (alpha × normalized_message_count) + (beta × normalized_delay)
	msgComponent := alpha.Mul(decimal.NewFromFloat(normalizedMessageCount))
	delayComponent := beta.Mul(decimal.NewFromFloat(normalizedDelay))

	return msgComponent.Add(delayComponent)
}

// UpdatePriority recalculates and updates the priority score
func (s *ConversationService) UpdatePriority(ctx context.Context, conv *domain.ConversationRef) error {
	priority, err := s.CalculatePriority(ctx, conv.TenantID, conv)
	if err != nil {
		return err
	}

	conv.PriorityScore = priority
	conv.UpdatedAt = time.Now().UTC()

	return s.repos.ConversationRefs.Update(ctx, conv)
}

// ==================== Get Labels for Conversation ====================

func (s *ConversationService) GetLabels(ctx context.Context, conversationID uuid.UUID) ([]*domain.Label, error) {
	// TODO: Implement GetByConversationID in label repository
	// For now, return empty slice - labels will be added in Stage 8
	return []*domain.Label{}, nil
}

// ==================== Batch Priority Update ====================

// UpdatePrioritiesForTenant recalculates priorities for all QUEUED conversations
// This should be called when tenant weights change or as a background job
func (s *ConversationService) UpdatePrioritiesForTenant(ctx context.Context, tenantID uuid.UUID) error {
	state := domain.ConversationStateQueued
	conversations, err := s.repos.ConversationRefs.ListWithFilters(ctx, repository.ConversationFilters{
		TenantID: tenantID,
		State:    &state,
		Limit:    1000, // Process in batches
	})
	if err != nil {
		return err
	}

	tenant, err := s.repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}

	for _, conv := range conversations {
		priority := s.calculatePriorityWithWeights(conv, tenant.PriorityWeightAlpha, tenant.PriorityWeightBeta)
		conv.PriorityScore = priority
		conv.UpdatedAt = time.Now().UTC()

		if err := s.repos.ConversationRefs.Update(ctx, conv); err != nil {
			s.logger.Warn("Failed to update priority for conversation",
				zap.String("conversation_id", conv.ID.String()),
				zap.Error(err))
		}
	}

	s.logger.Info("Updated priorities for tenant",
		zap.String("tenant_id", tenantID.String()),
		zap.Int("count", len(conversations)))

	return nil
}
