package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ==================== Tenant ====================

type Tenant struct {
	ID                  uuid.UUID
	Name                string
	PriorityWeightAlpha decimal.Decimal
	PriorityWeightBeta  decimal.Decimal
	CreatedAt           time.Time
	UpdatedAt           time.Time
	UpdatedBy           *uuid.UUID
}

func NewTenant(name string, alpha, beta decimal.Decimal) *Tenant {
	now := time.Now().UTC()
	return &Tenant{
		ID:                  uuid.Must(uuid.NewV7()),
		Name:                name,
		PriorityWeightAlpha: alpha,
		PriorityWeightBeta:  beta,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// ==================== Inbox ====================

type Inbox struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PhoneNumber string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewInbox(tenantID uuid.UUID, phoneNumber, displayName string) *Inbox {
	now := time.Now().UTC()
	return &Inbox{
		ID:          uuid.Must(uuid.NewV7()),
		TenantID:    tenantID,
		PhoneNumber: phoneNumber,
		DisplayName: displayName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ==================== Operator ====================

type Operator struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Role      OperatorRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOperator(tenantID uuid.UUID, role OperatorRole) *Operator {
	now := time.Now().UTC()
	return &Operator{
		ID:        uuid.Must(uuid.NewV7()),
		TenantID:  tenantID,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ==================== OperatorInboxSubscription ====================

type OperatorInboxSubscription struct {
	ID         uuid.UUID
	OperatorID uuid.UUID
	InboxID    uuid.UUID
	CreatedAt  time.Time
}

func NewOperatorInboxSubscription(operatorID, inboxID uuid.UUID) *OperatorInboxSubscription {
	return &OperatorInboxSubscription{
		ID:         uuid.Must(uuid.NewV7()),
		OperatorID: operatorID,
		InboxID:    inboxID,
		CreatedAt:  time.Now().UTC(),
	}
}

// ==================== OperatorStatus ====================

type OperatorStatus struct {
	ID                 uuid.UUID
	OperatorID         uuid.UUID
	Status             OperatorStatusType
	LastStatusChangeAt time.Time
}

func NewOperatorStatus(operatorID uuid.UUID) *OperatorStatus {
	return &OperatorStatus{
		ID:                 uuid.Must(uuid.NewV7()),
		OperatorID:         operatorID,
		Status:             OperatorStatusOffline,
		LastStatusChangeAt: time.Now().UTC(),
	}
}

func (os *OperatorStatus) SetStatus(status OperatorStatusType) {
	os.Status = status
	os.LastStatusChangeAt = time.Now().UTC()
}

// ==================== ConversationRef ====================

type ConversationRef struct {
	ID                     uuid.UUID
	TenantID               uuid.UUID
	InboxID                uuid.UUID
	ExternalConversationID string
	CustomerPhoneNumber    string
	State                  ConversationState
	AssignedOperatorID     *uuid.UUID
	LastMessageAt          time.Time
	MessageCount           int32
	PriorityScore          decimal.Decimal
	CreatedAt              time.Time
	UpdatedAt              time.Time
	ResolvedAt             *time.Time
}

func NewConversationRef(
	tenantID, inboxID uuid.UUID,
	externalID, customerPhone string,
) *ConversationRef {
	now := time.Now().UTC()
	return &ConversationRef{
		ID:                     uuid.Must(uuid.NewV7()),
		TenantID:               tenantID,
		InboxID:                inboxID,
		ExternalConversationID: externalID,
		CustomerPhoneNumber:    customerPhone,
		State:                  ConversationStateQueued,
		LastMessageAt:          now,
		MessageCount:           0,
		PriorityScore:          decimal.Zero,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
}

// Allocate assigns conversation to an operator
func (c *ConversationRef) Allocate(operatorID uuid.UUID) error {
	if !c.State.CanTransitionTo(ConversationStateAllocated) {
		return ErrInvalidStateTransition
	}
	c.State = ConversationStateAllocated
	c.AssignedOperatorID = &operatorID
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// Deallocate returns conversation to queue
func (c *ConversationRef) Deallocate() error {
	if !c.State.CanTransitionTo(ConversationStateQueued) {
		return ErrInvalidStateTransition
	}
	c.State = ConversationStateQueued
	c.AssignedOperatorID = nil
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// Resolve marks conversation as resolved
func (c *ConversationRef) Resolve() error {
	if !c.State.CanTransitionTo(ConversationStateResolved) {
		return ErrInvalidStateTransition
	}
	now := time.Now().UTC()
	c.State = ConversationStateResolved
	c.ResolvedAt = &now
	c.UpdatedAt = now
	return nil
}

// ==================== Label ====================

type Label struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	InboxID   uuid.UUID
	Name      string
	Color     *string
	CreatedBy *uuid.UUID
	CreatedAt time.Time
}

func NewLabel(tenantID, inboxID uuid.UUID, name string, color *string, createdBy *uuid.UUID) *Label {
	return &Label{
		ID:        uuid.Must(uuid.NewV7()),
		TenantID:  tenantID,
		InboxID:   inboxID,
		Name:      name,
		Color:     color,
		CreatedBy: createdBy,
		CreatedAt: time.Now().UTC(),
	}
}

// ==================== ConversationLabel ====================

type ConversationLabel struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	LabelID        uuid.UUID
	CreatedAt      time.Time
}

func NewConversationLabel(conversationID, labelID uuid.UUID) *ConversationLabel {
	return &ConversationLabel{
		ID:             uuid.Must(uuid.NewV7()),
		ConversationID: conversationID,
		LabelID:        labelID,
		CreatedAt:      time.Now().UTC(),
	}
}

// ==================== GracePeriodAssignment ====================

type GracePeriodAssignment struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	OperatorID     uuid.UUID
	ExpiresAt      time.Time
	Reason         GracePeriodReason
	CreatedAt      time.Time
}

func NewGracePeriodAssignment(
	conversationID, operatorID uuid.UUID,
	expiresAt time.Time,
	reason GracePeriodReason,
) *GracePeriodAssignment {
	return &GracePeriodAssignment{
		ID:             uuid.Must(uuid.NewV7()),
		ConversationID: conversationID,
		OperatorID:     operatorID,
		ExpiresAt:      expiresAt,
		Reason:         reason,
		CreatedAt:      time.Now().UTC(),
	}
}

func (g *GracePeriodAssignment) IsExpired() bool {
	return time.Now().UTC().After(g.ExpiresAt)
}
