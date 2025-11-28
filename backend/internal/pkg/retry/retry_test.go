package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDo_Success(t *testing.T) {
	cfg := Config{MaxAttempts: 3, InitialBackoff: 10 * time.Millisecond}

	calls := 0
	err := Do(context.Background(), cfg, func() error {
		calls++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestDo_RetryUntilSuccess(t *testing.T) {
	cfg := Config{MaxAttempts: 3, InitialBackoff: 10 * time.Millisecond}

	calls := 0
	err := Do(context.Background(), cfg, func() error {
		calls++
		if calls < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, calls)
}

func TestDo_MaxAttemptsExceeded(t *testing.T) {
	cfg := Config{MaxAttempts: 3, InitialBackoff: 10 * time.Millisecond}

	calls := 0
	err := Do(context.Background(), cfg, func() error {
		calls++
		return errors.New("always fails")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max retries")
	assert.Equal(t, 3, calls)
}

func TestDo_ContextCancelled(t *testing.T) {
	cfg := Config{MaxAttempts: 10, InitialBackoff: 1 * time.Second}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := Do(ctx, cfg, func() error {
		return errors.New("keep failing")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "retry cancelled")
}

func TestDo_PermanentError(t *testing.T) {
	cfg := Config{MaxAttempts: 5, InitialBackoff: 10 * time.Millisecond}

	calls := 0
	err := Do(context.Background(), cfg, func() error {
		calls++
		return MarkPermanent(errors.New("permanent error"))
	})

	assert.Error(t, err)
	assert.Equal(t, 1, calls) // Should not retry permanent errors
	assert.True(t, IsPermanent(err))
}

func TestDoWithResult_Success(t *testing.T) {
	cfg := Config{MaxAttempts: 3, InitialBackoff: 10 * time.Millisecond}

	result, err := DoWithResult(context.Background(), cfg, func() (string, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestDoWithResult_RetryUntilSuccess(t *testing.T) {
	cfg := Config{MaxAttempts: 3, InitialBackoff: 10 * time.Millisecond}

	calls := 0
	result, err := DoWithResult(context.Background(), cfg, func() (int, error) {
		calls++
		if calls < 3 {
			return 0, errors.New("temporary error")
		}
		return 42, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 3, calls)
}
