package security

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithRetry_ImmediateSuccess(t *testing.T) {
	calls := 0
	err := withRetry(context.Background(), 3, func() error {
		calls++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestWithRetry_TransientThenSuccess(t *testing.T) {
	calls := 0
	err := withRetry(context.Background(), 3, func() error {
		calls++
		if calls < 3 {
			return ErrKMSThrottled
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 3, calls)
}

func TestWithRetry_NonTransientError(t *testing.T) {
	calls := 0
	err := withRetry(context.Background(), 3, func() error {
		calls++
		return ErrKMSAccessDenied
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrKMSAccessDenied))
	assert.Equal(t, 1, calls, "non-transient errors should not be retried")
}

func TestWithRetry_ExhaustsRetries(t *testing.T) {
	calls := 0
	err := withRetry(context.Background(), 2, func() error {
		calls++
		return ErrKMSUnavailable
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrKMSUnavailable))
	assert.Equal(t, 3, calls, "should attempt 1 + 2 retries")
}

func TestWithRetry_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	err := withRetry(ctx, 5, func() error {
		calls++
		return ErrKMSThrottled
	})
	// First call happens, then context cancel prevents retries.
	assert.Error(t, err)
	assert.LessOrEqual(t, calls, 2)
}

func TestIsTransient(t *testing.T) {
	tests := []struct {
		give error
		want bool
	}{
		{ErrKMSUnavailable, true},
		{ErrKMSThrottled, true},
		{ErrKMSAccessDenied, false},
		{ErrKMSKeyDisabled, false},
		{ErrKMSInvalidKey, false},
		{errors.New("random error"), false},
		{&KMSError{Provider: "aws", Op: "encrypt", KeyID: "k1", Err: ErrKMSThrottled}, true},
	}

	for _, tt := range tests {
		t.Run(tt.give.Error(), func(t *testing.T) {
			assert.Equal(t, tt.want, IsTransient(tt.give))
		})
	}
}
