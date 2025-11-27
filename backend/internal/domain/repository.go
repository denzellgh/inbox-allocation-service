package domain

import (
	"context"

	"github.com/google/uuid"
)

// ==================== TenantRepository ====================

type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetByName(ctx context.Context, name string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ==================== InboxRepository ====================

type InboxRepository interface {
	Create(ctx context.Context, inbox *Inbox) error
	GetByID(ctx context.Context, id uuid.UUID) (*Inbox, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*Inbox, error)
	GetByPhoneNumber(ctx context.Context, tenantID uuid.UUID, phoneNumber string) (*Inbox, error)
	Update(ctx context.Context, inbox *Inbox) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ==================== OperatorRepository ====================

type OperatorRepository interface {
	Create(ctx context.Context, operator *Operator) error
	GetByID(ctx context.Context, id uuid.UUID) (*Operator, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*Operator, error)
	GetByTenantAndRole(ctx context.Context, tenantID uuid.UUID, role OperatorRole) ([]*Operator, error)
	Update(ctx context.Context, operator *Operator) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ==================== OperatorInboxSubscriptionRepository ====================

type OperatorInboxSubscriptionRepository interface {
	Create(ctx context.Context, subscription *OperatorInboxSubscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*OperatorInboxSubscription, error)
	GetByOperatorID(ctx context.Context, operatorID uuid.UUID) ([]*OperatorInboxSubscription, error)
	GetByInboxID(ctx context.Context, inboxID uuid.UUID) ([]*OperatorInboxSubscription, error)
	GetByOperatorAndInbox(ctx context.Context, operatorID, inboxID uuid.UUID) (*OperatorInboxSubscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByOperatorAndInbox(ctx context.Context, operatorID, inboxID uuid.UUID) error
	// Returns list of inbox IDs the operator is subscribed to
	GetSubscribedInboxIDs(ctx context.Context, operatorID uuid.UUID) ([]uuid.UUID, error)
	// Check if operator is subscribed to a specific inbox
	IsSubscribed(ctx context.Context, operatorID, inboxID uuid.UUID) (bool, error)
}

// ==================== OperatorStatusRepository ====================

type OperatorStatusRepository interface {
	Create(ctx context.Context, status *OperatorStatus) error
	GetByOperatorID(ctx context.Context, operatorID uuid.UUID) (*OperatorStatus, error)
	Update(ctx context.Context, status *OperatorStatus) error
	GetAvailableOperators(ctx context.Context, tenantID uuid.UUID) ([]*OperatorStatus, error)
}

// ==================== ConversationRefRepository ====================

type ConversationFilter struct {
	TenantID           uuid.UUID
	State              *ConversationState
	InboxID            *uuid.UUID
	AssignedOperatorID *uuid.UUID
	LabelID            *uuid.UUID
	Limit              int
	Cursor             *uuid.UUID // For pagination
}

type ConversationRefRepository interface {
	Create(ctx context.Context, conv *ConversationRef) error
	GetByID(ctx context.Context, id uuid.UUID) (*ConversationRef, error)
	GetByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*ConversationRef, error)
	GetByFilter(ctx context.Context, filter ConversationFilter) ([]*ConversationRef, error)
	SearchByPhone(ctx context.Context, tenantID uuid.UUID, phoneNumber string) ([]*ConversationRef, error)
	Update(ctx context.Context, conv *ConversationRef) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Allocation-specific methods (with locking)
	// Returns the next available conversation for allocation using FOR UPDATE SKIP LOCKED
	GetNextForAllocation(ctx context.Context, tenantID uuid.UUID, inboxIDs []uuid.UUID, limit int) ([]*ConversationRef, error)
	// Lock a specific conversation for claim
	LockForClaim(ctx context.Context, id uuid.UUID) (*ConversationRef, error)

	// Bulk operations
	GetByOperatorID(ctx context.Context, tenantID, operatorID uuid.UUID, state *ConversationState) ([]*ConversationRef, error)
}

// ==================== LabelRepository ====================

type LabelRepository interface {
	Create(ctx context.Context, label *Label) error
	GetByID(ctx context.Context, id uuid.UUID) (*Label, error)
	GetByInboxID(ctx context.Context, tenantID, inboxID uuid.UUID) ([]*Label, error)
	GetByName(ctx context.Context, inboxID uuid.UUID, name string) (*Label, error)
	Update(ctx context.Context, label *Label) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ==================== ConversationLabelRepository ====================

type ConversationLabelRepository interface {
	Create(ctx context.Context, cl *ConversationLabel) error
	GetByConversationID(ctx context.Context, conversationID uuid.UUID) ([]*ConversationLabel, error)
	GetByLabelID(ctx context.Context, labelID uuid.UUID) ([]*ConversationLabel, error)
	Delete(ctx context.Context, conversationID, labelID uuid.UUID) error
	DeleteAllForConversation(ctx context.Context, conversationID uuid.UUID) error
	Exists(ctx context.Context, conversationID, labelID uuid.UUID) (bool, error)
}

// ==================== GracePeriodAssignmentRepository ====================

type GracePeriodAssignmentRepository interface {
	Create(ctx context.Context, gpa *GracePeriodAssignment) error
	GetByConversationID(ctx context.Context, conversationID uuid.UUID) (*GracePeriodAssignment, error)
	GetByOperatorID(ctx context.Context, operatorID uuid.UUID) ([]*GracePeriodAssignment, error)
	GetExpired(ctx context.Context, limit int) ([]*GracePeriodAssignment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByOperatorID(ctx context.Context, operatorID uuid.UUID) error
	DeleteByConversationID(ctx context.Context, conversationID uuid.UUID) error

	// For worker: get and lock expired assignments
	GetAndLockExpired(ctx context.Context, limit int) ([]*GracePeriodAssignment, error)
}

// ==================== IdempotencyRepository ====================

// IdempotencyRepository handles idempotency key storage
type IdempotencyRepository interface {
	// Create stores a new idempotency key
	Create(ctx context.Context, ik *IdempotencyKey) error

	// GetByKey retrieves an idempotency key by tenant and key
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*IdempotencyKey, error)

	// Delete removes an idempotency key
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteExpired removes all expired idempotency keys
	DeleteExpired(ctx context.Context) (int64, error)

	// GetExpiredForCleanup gets expired keys with lock for distributed cleanup
	GetExpiredForCleanup(ctx context.Context, limit int) ([]*IdempotencyKey, error)
}
