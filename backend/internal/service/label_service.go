package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/pkg/logger"
	"github.com/inbox-allocation-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	ErrLabelNotFound         = errors.New("label not found")
	ErrLabelNameConflict     = errors.New("label name already exists in this inbox")
	ErrLabelInboxMismatch    = errors.New("label inbox does not match conversation inbox")
	ErrLabelPermissionDenied = errors.New("insufficient permissions for label operation")
)

type LabelService struct {
	repos  *repository.RepositoryContainer
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewLabelService(repos *repository.RepositoryContainer, pool *pgxpool.Pool, log *logger.Logger) *LabelService {
	return &LabelService{
		repos:  repos,
		pool:   pool,
		logger: log,
	}
}

// ==================== Create Label ====================

// CreateLabel creates a new label for an inbox
// Permission: Manager or Admin only
func (s *LabelService) CreateLabel(
	ctx context.Context,
	tenantID, operatorID, inboxID uuid.UUID,
	role domain.OperatorRole,
	name string,
	color *string,
) (*domain.Label, error) {
	start := time.Now()

	// Check permissions
	if !s.canManageLabels(role) {
		return nil, ErrLabelPermissionDenied
	}

	// Verify inbox exists and belongs to tenant
	inbox, err := s.repos.Inboxes.GetByID(ctx, inboxID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if inbox.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}

	// Check for duplicate name in inbox
	name = strings.TrimSpace(name)
	existing, err := s.repos.Labels.GetByName(ctx, inboxID, name)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrLabelNameConflict
	}

	// Create label
	label := domain.NewLabel(tenantID, inboxID, name, color, &operatorID)

	if err := s.repos.Labels.Create(ctx, label); err != nil {
		return nil, err
	}

	s.logger.Info("Label created",
		zap.String("label_id", label.ID.String()),
		zap.String("inbox_id", inboxID.String()),
		zap.String("name", name),
		zap.String("created_by", operatorID.String()),
		zap.Duration("duration", time.Since(start)))

	return label, nil
}

// ==================== Update Label ====================

// UpdateLabel updates an existing label
// Permission: Manager or Admin only
func (s *LabelService) UpdateLabel(
	ctx context.Context,
	tenantID, operatorID, labelID uuid.UUID,
	role domain.OperatorRole,
	name *string,
	color *string,
) (*domain.Label, error) {
	start := time.Now()

	// Check permissions
	if !s.canManageLabels(role) {
		return nil, ErrLabelPermissionDenied
	}

	// Get existing label
	label, err := s.repos.Labels.GetByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrLabelNotFound
		}
		return nil, err
	}

	// Verify tenant
	if label.TenantID != tenantID {
		return nil, ErrLabelNotFound
	}

	// Update fields
	if name != nil {
		newName := strings.TrimSpace(*name)
		// Check for duplicate if name changed
		if newName != label.Name {
			existing, err := s.repos.Labels.GetByName(ctx, label.InboxID, newName)
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return nil, err
			}
			if existing != nil && existing.ID != labelID {
				return nil, ErrLabelNameConflict
			}
			label.Name = newName
		}
	}

	if color != nil {
		label.Color = color
	}

	if err := s.repos.Labels.Update(ctx, label); err != nil {
		return nil, err
	}

	s.logger.Info("Label updated",
		zap.String("label_id", labelID.String()),
		zap.String("updated_by", operatorID.String()),
		zap.Duration("duration", time.Since(start)))

	return label, nil
}

// ==================== Delete Label ====================

// DeleteLabel deletes a label
// Permission: Manager or Admin only
func (s *LabelService) DeleteLabel(
	ctx context.Context,
	tenantID, operatorID, labelID uuid.UUID,
	role domain.OperatorRole,
) error {
	start := time.Now()

	// Check permissions
	if !s.canManageLabels(role) {
		return ErrLabelPermissionDenied
	}

	// Get existing label
	label, err := s.repos.Labels.GetByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrLabelNotFound
		}
		return err
	}

	// Verify tenant
	if label.TenantID != tenantID {
		return ErrLabelNotFound
	}

	// Delete label (cascade deletes conversation_labels via DB constraint)
	if err := s.repos.Labels.Delete(ctx, labelID); err != nil {
		return err
	}

	s.logger.Info("Label deleted",
		zap.String("label_id", labelID.String()),
		zap.String("inbox_id", label.InboxID.String()),
		zap.String("deleted_by", operatorID.String()),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// ==================== List Labels ====================

// ListLabelsByInbox lists all labels for an inbox
// Permission: Subscribed Operator, Manager, or Admin
func (s *LabelService) ListLabelsByInbox(
	ctx context.Context,
	tenantID, operatorID, inboxID uuid.UUID,
	role domain.OperatorRole,
) ([]*domain.Label, error) {
	// Verify inbox exists and belongs to tenant
	inbox, err := s.repos.Inboxes.GetByID(ctx, inboxID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if inbox.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}

	// Check permissions for operators
	if role == domain.OperatorRoleOperator {
		isSubscribed, err := s.repos.Subscriptions.IsSubscribed(ctx, operatorID, inboxID)
		if err != nil {
			return nil, err
		}
		if !isSubscribed {
			return nil, ErrLabelPermissionDenied
		}
	}

	return s.repos.Labels.GetByInboxID(ctx, tenantID, inboxID)
}

// ==================== Attach Label ====================

// AttachLabelToConversation attaches a label to a conversation
// Permission: Subscribed Operator, Manager, or Admin
// Idempotent: If already attached, returns success
func (s *LabelService) AttachLabelToConversation(
	ctx context.Context,
	tenantID, operatorID, conversationID, labelID uuid.UUID,
	role domain.OperatorRole,
) error {
	start := time.Now()

	// Begin transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get conversation
	conv, err := s.repos.ConversationRefs.GetByID(ctx, conversationID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	// Verify tenant
	if conv.TenantID != tenantID {
		return domain.ErrNotFound
	}

	// Get label
	label, err := s.repos.Labels.GetByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrLabelNotFound
		}
		return err
	}

	// Verify label belongs to same tenant
	if label.TenantID != tenantID {
		return ErrLabelNotFound
	}

	// Verify label inbox matches conversation inbox
	if label.InboxID != conv.InboxID {
		return ErrLabelInboxMismatch
	}

	// Check permissions for operators
	if role == domain.OperatorRoleOperator {
		isSubscribed, err := s.repos.Subscriptions.IsSubscribed(ctx, operatorID, conv.InboxID)
		if err != nil {
			return err
		}
		if !isSubscribed {
			return ErrLabelPermissionDenied
		}
	}

	// Check if already attached (idempotency)
	exists, err := s.repos.ConversationLabels.Exists(ctx, conversationID, labelID)
	if err != nil {
		return err
	}
	if exists {
		s.logger.Debug("Label already attached to conversation",
			zap.String("conversation_id", conversationID.String()),
			zap.String("label_id", labelID.String()))
		return nil
	}

	// Create association
	cl := domain.NewConversationLabel(conversationID, labelID)
	if err := s.repos.ConversationLabels.Create(ctx, cl); err != nil {
		return err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.logger.Info("Label attached to conversation",
		zap.String("conversation_id", conversationID.String()),
		zap.String("label_id", labelID.String()),
		zap.String("attached_by", operatorID.String()),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// ==================== Detach Label ====================

// DetachLabelFromConversation detaches a label from a conversation
// Permission: Subscribed Operator, Manager, or Admin
// Idempotent: If not attached, returns success
func (s *LabelService) DetachLabelFromConversation(
	ctx context.Context,
	tenantID, operatorID, conversationID, labelID uuid.UUID,
	role domain.OperatorRole,
) error {
	start := time.Now()

	// Get conversation
	conv, err := s.repos.ConversationRefs.GetByID(ctx, conversationID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	// Verify tenant
	if conv.TenantID != tenantID {
		return domain.ErrNotFound
	}

	// Get label
	label, err := s.repos.Labels.GetByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrLabelNotFound
		}
		return err
	}

	// Verify label belongs to same tenant
	if label.TenantID != tenantID {
		return ErrLabelNotFound
	}

	// Check permissions for operators
	if role == domain.OperatorRoleOperator {
		isSubscribed, err := s.repos.Subscriptions.IsSubscribed(ctx, operatorID, conv.InboxID)
		if err != nil {
			return err
		}
		if !isSubscribed {
			return ErrLabelPermissionDenied
		}
	}

	// Check if exists (idempotency)
	exists, err := s.repos.ConversationLabels.Exists(ctx, conversationID, labelID)
	if err != nil {
		return err
	}
	if !exists {
		s.logger.Debug("Label not attached to conversation",
			zap.String("conversation_id", conversationID.String()),
			zap.String("label_id", labelID.String()))
		return nil
	}

	// Delete association
	if err := s.repos.ConversationLabels.Delete(ctx, conversationID, labelID); err != nil {
		return err
	}

	s.logger.Info("Label detached from conversation",
		zap.String("conversation_id", conversationID.String()),
		zap.String("label_id", labelID.String()),
		zap.String("detached_by", operatorID.String()),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// ==================== Permission Helpers ====================

// canManageLabels checks if caller can create/update/delete labels
func (s *LabelService) canManageLabels(role domain.OperatorRole) bool {
	return role == domain.OperatorRoleAdmin || role == domain.OperatorRoleManager
}
