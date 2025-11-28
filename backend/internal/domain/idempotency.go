package domain

import (
	"time"

	"github.com/google/uuid"
)

// IdempotencyKey represents a stored idempotency key with its response
type IdempotencyKey struct {
	ID             uuid.UUID
	Key            string
	TenantID       uuid.UUID
	Endpoint       string
	Method         string
	RequestHash    *string
	ResponseStatus int
	ResponseBody   []byte
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

// NewIdempotencyKey creates a new idempotency key record
func NewIdempotencyKey(
	key string,
	tenantID uuid.UUID,
	endpoint, method string,
	requestHash *string,
	responseStatus int,
	responseBody []byte,
	ttl time.Duration,
) *IdempotencyKey {
	now := time.Now().UTC()
	return &IdempotencyKey{
		ID:             uuid.Must(uuid.NewV7()),
		Key:            key,
		TenantID:       tenantID,
		Endpoint:       endpoint,
		Method:         method,
		RequestHash:    requestHash,
		ResponseStatus: responseStatus,
		ResponseBody:   responseBody,
		CreatedAt:      now,
		ExpiresAt:      now.Add(ttl),
	}
}

// IsExpired checks if the idempotency key has expired
func (ik *IdempotencyKey) IsExpired() bool {
	return time.Now().UTC().After(ik.ExpiresAt)
}

// DefaultIdempotencyTTL is the default time-to-live for idempotency keys
const DefaultIdempotencyTTL = 24 * time.Hour
