package domain

import (
	"errors"

	"github.com/google/uuid"
)

// ==================== ConversationState ====================

type ConversationState string

const (
	ConversationStateQueued    ConversationState = "QUEUED"
	ConversationStateAllocated ConversationState = "ALLOCATED"
	ConversationStateResolved  ConversationState = "RESOLVED"
)

func (s ConversationState) IsValid() bool {
	switch s {
	case ConversationStateQueued, ConversationStateAllocated, ConversationStateResolved:
		return true
	}
	return false
}

func (s ConversationState) String() string {
	return string(s)
}

// CanTransitionTo validates state transitions
func (s ConversationState) CanTransitionTo(target ConversationState) bool {
	transitions := map[ConversationState][]ConversationState{
		ConversationStateQueued:    {ConversationStateAllocated},
		ConversationStateAllocated: {ConversationStateQueued, ConversationStateResolved},
		ConversationStateResolved:  {}, // Terminal state
	}
	for _, allowed := range transitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

// ==================== OperatorRole ====================

type OperatorRole string

const (
	OperatorRoleOperator OperatorRole = "OPERATOR"
	OperatorRoleManager  OperatorRole = "MANAGER"
	OperatorRoleAdmin    OperatorRole = "ADMIN"
)

func (r OperatorRole) IsValid() bool {
	switch r {
	case OperatorRoleOperator, OperatorRoleManager, OperatorRoleAdmin:
		return true
	}
	return false
}

func (r OperatorRole) String() string {
	return string(r)
}

// CanPerformAction checks role-based permissions
func (r OperatorRole) CanResolve() bool {
	return true // All roles can resolve their own conversations
}

func (r OperatorRole) CanDeallocate() bool {
	return r == OperatorRoleManager || r == OperatorRoleAdmin
}

func (r OperatorRole) CanReassign() bool {
	return r == OperatorRoleManager || r == OperatorRoleAdmin
}

func (r OperatorRole) CanMoveInbox() bool {
	return r == OperatorRoleManager || r == OperatorRoleAdmin
}

// ==================== OperatorStatusType ====================

type OperatorStatusType string

const (
	OperatorStatusAvailable OperatorStatusType = "AVAILABLE"
	OperatorStatusOffline   OperatorStatusType = "OFFLINE"
)

func (s OperatorStatusType) IsValid() bool {
	switch s {
	case OperatorStatusAvailable, OperatorStatusOffline:
		return true
	}
	return false
}

func (s OperatorStatusType) String() string {
	return string(s)
}

// ==================== GracePeriodReason ====================

type GracePeriodReason string

const (
	GracePeriodReasonOffline GracePeriodReason = "OFFLINE"
	GracePeriodReasonManual  GracePeriodReason = "MANUAL"
)

func (r GracePeriodReason) IsValid() bool {
	switch r {
	case GracePeriodReasonOffline, GracePeriodReasonManual:
		return true
	}
	return false
}

func (r GracePeriodReason) String() string {
	return string(r)
}

// ==================== TenantID (typed UUID) ====================

type TenantID uuid.UUID

func NewTenantID() TenantID {
	return TenantID(uuid.Must(uuid.NewV7()))
}

func ParseTenantID(s string) (TenantID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return TenantID{}, errors.New("invalid tenant ID format")
	}
	return TenantID(id), nil
}

func (id TenantID) String() string {
	return uuid.UUID(id).String()
}

func (id TenantID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id TenantID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

// ==================== OperatorID ====================

type OperatorID uuid.UUID

func NewOperatorID() OperatorID {
	return OperatorID(uuid.Must(uuid.NewV7()))
}

func ParseOperatorID(s string) (OperatorID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return OperatorID{}, errors.New("invalid operator ID format")
	}
	return OperatorID(id), nil
}

func (id OperatorID) String() string {
	return uuid.UUID(id).String()
}

func (id OperatorID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id OperatorID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

// ==================== InboxID ====================

type InboxID uuid.UUID

func NewInboxID() InboxID {
	return InboxID(uuid.Must(uuid.NewV7()))
}

func ParseInboxID(s string) (InboxID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return InboxID{}, errors.New("invalid inbox ID format")
	}
	return InboxID(id), nil
}

func (id InboxID) String() string {
	return uuid.UUID(id).String()
}

func (id InboxID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id InboxID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

// ==================== ConversationID ====================

type ConversationID uuid.UUID

func NewConversationID() ConversationID {
	return ConversationID(uuid.Must(uuid.NewV7()))
}

func ParseConversationID(s string) (ConversationID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ConversationID{}, errors.New("invalid conversation ID format")
	}
	return ConversationID(id), nil
}

func (id ConversationID) String() string {
	return uuid.UUID(id).String()
}

func (id ConversationID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id ConversationID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

// ==================== LabelID ====================

type LabelID uuid.UUID

func NewLabelID() LabelID {
	return LabelID(uuid.Must(uuid.NewV7()))
}

func ParseLabelID(s string) (LabelID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return LabelID{}, errors.New("invalid label ID format")
	}
	return LabelID(id), nil
}

func (id LabelID) String() string {
	return uuid.UUID(id).String()
}

func (id LabelID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id LabelID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

// ==================== SubscriptionID ====================

type SubscriptionID uuid.UUID

func NewSubscriptionID() SubscriptionID {
	return SubscriptionID(uuid.Must(uuid.NewV7()))
}

func ParseSubscriptionID(s string) (SubscriptionID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return SubscriptionID{}, errors.New("invalid subscription ID format")
	}
	return SubscriptionID(id), nil
}

func (id SubscriptionID) String() string {
	return uuid.UUID(id).String()
}

func (id SubscriptionID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id SubscriptionID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

// ==================== GracePeriodID ====================

type GracePeriodID uuid.UUID

func NewGracePeriodID() GracePeriodID {
	return GracePeriodID(uuid.Must(uuid.NewV7()))
}

func ParseGracePeriodID(s string) (GracePeriodID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return GracePeriodID{}, errors.New("invalid grace period ID format")
	}
	return GracePeriodID(id), nil
}

func (id GracePeriodID) String() string {
	return uuid.UUID(id).String()
}

func (id GracePeriodID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id GracePeriodID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}
