package testutil

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertUUIDNotZero asserts that a UUID is not zero
func AssertUUIDNotZero(t *testing.T, id uuid.UUID, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotEqual(t, uuid.Nil, id, msgAndArgs...)
}

// RequireUUIDNotZero requires that a UUID is not zero
func RequireUUIDNotZero(t *testing.T, id uuid.UUID, msgAndArgs ...interface{}) {
	t.Helper()
	require.NotEqual(t, uuid.Nil, id, msgAndArgs...)
}

// AssertNoError asserts no error with a message
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	assert.NoError(t, err, msg)
}

// RequireNoError requires no error with a message
func RequireNoError(t *testing.T, err error, msg string) {
	t.Helper()
	require.NoError(t, err, msg)
}
