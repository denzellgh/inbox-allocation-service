package retry

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// Config defines retry behavior
type Config struct {
	// MaxAttempts is the maximum number of attempts (including first try)
	MaxAttempts int
	// InitialBackoff is the initial backoff duration
	InitialBackoff time.Duration
	// MaxBackoff is the maximum backoff duration
	MaxBackoff time.Duration
	// BackoffFactor is the multiplier for exponential backoff
	BackoffFactor float64
	// Jitter adds randomness to backoff (0.0 to 1.0)
	Jitter float64
	// RetryableErrors are errors that should trigger a retry
	RetryableErrors []error
	// OnRetry is called before each retry attempt
	OnRetry func(attempt int, err error, nextBackoff time.Duration)
}

// DefaultConfig returns sensible defaults for retry
func DefaultConfig() Config {
	return Config{
		MaxAttempts:    3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		BackoffFactor:  2.0,
		Jitter:         0.1,
	}
}

// Do executes fn with retries according to config
func Do(ctx context.Context, cfg Config, fn func() error) error {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}

	var lastErr error
	backoff := cfg.InitialBackoff

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Execute function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is permanent (non-retryable)
		if IsPermanent(err) {
			return err
		}

		// Check if error is retryable
		if !isRetryable(err, cfg.RetryableErrors) {
			return err
		}

		// Last attempt, don't wait
		if attempt == cfg.MaxAttempts {
			break
		}

		// Calculate backoff with jitter
		sleep := calculateBackoff(backoff, cfg.MaxBackoff, cfg.Jitter)

		// Call OnRetry callback
		if cfg.OnRetry != nil {
			cfg.OnRetry(attempt, err, sleep)
		}

		// Wait or context cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(sleep):
		}

		// Increase backoff for next attempt
		backoff = time.Duration(float64(backoff) * cfg.BackoffFactor)
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxAttempts, lastErr)
}

// DoWithResult executes fn with retries and returns the result
func DoWithResult[T any](ctx context.Context, cfg Config, fn func() (T, error)) (T, error) {
	var result T
	err := Do(ctx, cfg, func() error {
		var err error
		result, err = fn()
		return err
	})
	return result, err
}

// isRetryable checks if error should trigger a retry
func isRetryable(err error, retryableErrors []error) bool {
	// If no specific errors defined, retry all
	if len(retryableErrors) == 0 {
		return true
	}

	for _, retryable := range retryableErrors {
		if errors.Is(err, retryable) {
			return true
		}
	}
	return false
}

// calculateBackoff calculates backoff with jitter
func calculateBackoff(backoff, maxBackoff time.Duration, jitter float64) time.Duration {
	// Apply jitter
	if jitter > 0 {
		jitterAmount := float64(backoff) * jitter * (rand.Float64()*2 - 1)
		backoff = backoff + time.Duration(jitterAmount)
	}

	// Cap at max
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	// Ensure positive
	if backoff < 0 {
		backoff = 0
	}

	return backoff
}

// Permanent wraps an error to indicate it should not be retried
type Permanent struct {
	Err error
}

func (e Permanent) Error() string {
	return e.Err.Error()
}

func (e Permanent) Unwrap() error {
	return e.Err
}

// IsPermanent checks if error is permanent
func IsPermanent(err error) bool {
	var permanent Permanent
	return errors.As(err, &permanent)
}

// MarkPermanent marks an error as permanent (non-retryable)
func MarkPermanent(err error) error {
	return Permanent{Err: err}
}
