package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/pkg/logger"
	"github.com/inbox-allocation-service/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrIdempotencyKeyExists   = errors.New("idempotency key already exists")
	ErrIdempotencyKeyNotFound = errors.New("idempotency key not found")
	ErrIdempotencyKeyExpired  = errors.New("idempotency key has expired")
	ErrRequestHashMismatch    = errors.New("request body does not match stored hash")
)

// IdempotencyConfig holds configuration for idempotency
type IdempotencyConfig struct {
	TTL             time.Duration
	CleanupInterval time.Duration
	CleanupBatch    int
}

// DefaultIdempotencyConfig returns sensible defaults
func DefaultIdempotencyConfig() IdempotencyConfig {
	return IdempotencyConfig{
		TTL:             24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
		CleanupBatch:    100,
	}
}

type IdempotencyService struct {
	repos  *repository.RepositoryContainer
	config IdempotencyConfig
	logger *logger.Logger
}

func NewIdempotencyService(
	repos *repository.RepositoryContainer,
	config IdempotencyConfig,
	log *logger.Logger,
) *IdempotencyService {
	return &IdempotencyService{
		repos:  repos,
		config: config,
		logger: log,
	}
}

// CachedResponse holds a cached response from an idempotency key
type CachedResponse struct {
	Status int
	Body   []byte
}

// CheckKey checks if an idempotency key exists and returns the cached response
// Returns nil if key doesn't exist (proceed with request)
// Returns CachedResponse if key exists (return cached response)
// Returns error if key exists but request hash doesn't match
func (s *IdempotencyService) CheckKey(
	ctx context.Context,
	tenantID uuid.UUID,
	key string,
	requestBody []byte,
) (*CachedResponse, error) {
	ik, err := s.repos.Idempotency.GetByKey(ctx, tenantID, key)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Key doesn't exist, proceed with request
			return nil, nil
		}
		return nil, err
	}

	// Key exists - check if expired
	if ik.IsExpired() {
		// Delete expired key and proceed
		_ = s.repos.Idempotency.Delete(ctx, ik.ID)
		return nil, nil
	}

	// Key exists and not expired
	// Optionally validate request hash if provided
	if ik.RequestHash != nil && len(requestBody) > 0 {
		hash := hashRequestBody(requestBody)
		if hash != *ik.RequestHash {
			s.logger.Warn("Idempotency key reused with different request body",
				zap.String("key", key),
				zap.String("tenant_id", tenantID.String()))
			return nil, ErrRequestHashMismatch
		}
	}

	s.logger.Info("Returning cached response for idempotency key",
		zap.String("key", key),
		zap.String("tenant_id", tenantID.String()),
		zap.Int("status", ik.ResponseStatus))

	return &CachedResponse{
		Status: ik.ResponseStatus,
		Body:   ik.ResponseBody,
	}, nil
}

// StoreResult stores the result of a request with an idempotency key
func (s *IdempotencyService) StoreResult(
	ctx context.Context,
	tenantID uuid.UUID,
	key string,
	endpoint, method string,
	requestBody []byte,
	responseStatus int,
	responseBody []byte,
) error {
	var requestHash *string
	if len(requestBody) > 0 {
		h := hashRequestBody(requestBody)
		requestHash = &h
	}

	ik := domain.NewIdempotencyKey(
		key,
		tenantID,
		endpoint,
		method,
		requestHash,
		responseStatus,
		responseBody,
		s.config.TTL,
	)

	if err := s.repos.Idempotency.Create(ctx, ik); err != nil {
		s.logger.Error("Failed to store idempotency key",
			zap.String("key", key),
			zap.Error(err))
		return err
	}

	s.logger.Debug("Stored idempotency key",
		zap.String("key", key),
		zap.String("tenant_id", tenantID.String()),
		zap.Int("status", responseStatus),
		zap.Time("expires_at", ik.ExpiresAt))

	return nil
}

// CleanupExpired removes expired idempotency keys
func (s *IdempotencyService) CleanupExpired(ctx context.Context) (int64, error) {
	count, err := s.repos.Idempotency.DeleteExpired(ctx)
	if err != nil {
		return 0, err
	}

	if count > 0 {
		s.logger.Info("Cleaned up expired idempotency keys",
			zap.Int64("count", count))
	}

	return count, nil
}

// hashRequestBody creates a SHA256 hash of the request body
func hashRequestBody(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}
