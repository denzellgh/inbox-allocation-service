package domain

import "errors"

var (
	// Entity errors
	ErrNotFound               = errors.New("entity not found")
	ErrAlreadyExists          = errors.New("entity already exists")
	ErrInvalidStateTransition = errors.New("invalid state transition")

	// Validation errors
	ErrInvalidTenantID       = errors.New("invalid tenant ID")
	ErrInvalidOperatorID     = errors.New("invalid operator ID")
	ErrInvalidInboxID        = errors.New("invalid inbox ID")
	ErrInvalidConversationID = errors.New("invalid conversation ID")
	ErrInvalidLabelID        = errors.New("invalid label ID")

	// Business logic errors
	ErrOperatorNotSubscribed       = errors.New("operator not subscribed to inbox")
	ErrOperatorNotAvailable        = errors.New("operator is not available")
	ErrConversationNotQueued       = errors.New("conversation is not in queued state")
	ErrConversationAlreadyAssigned = errors.New("conversation already assigned")
	ErrNoConversationsAvailable    = errors.New("no conversations available for allocation")
	ErrInsufficientPermissions     = errors.New("insufficient permissions for this operation")

	// Concurrency errors
	ErrConcurrentModification = errors.New("concurrent modification detected")
	ErrLockAcquisitionFailed  = errors.New("failed to acquire lock")
)
