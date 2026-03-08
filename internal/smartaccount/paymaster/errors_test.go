package paymaster

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTransient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		err  error
		want bool
	}{
		{
			give: "ErrPaymasterTimeout is transient",
			err:  ErrPaymasterTimeout,
			want: true,
		},
		{
			give: "wrapped ErrPaymasterTimeout is transient",
			err:  fmt.Errorf("sponsor op: %w", ErrPaymasterTimeout),
			want: true,
		},
		{
			give: "double-wrapped ErrPaymasterTimeout is transient",
			err:  fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrPaymasterTimeout)),
			want: true,
		},
		{
			give: "ErrPaymasterRejected is not transient",
			err:  ErrPaymasterRejected,
			want: false,
		},
		{
			give: "ErrInsufficientToken is not transient",
			err:  ErrInsufficientToken,
			want: false,
		},
		{
			give: "ErrPaymasterNotConfigured is not transient",
			err:  ErrPaymasterNotConfigured,
			want: false,
		},
		{
			give: "unrelated error is not transient",
			err:  errors.New("something else"),
			want: false,
		},
		{
			give: "nil error is not transient",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsTransient(tt.err))
		})
	}
}

func TestIsPermanent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		err  error
		want bool
	}{
		{
			give: "ErrPaymasterRejected is permanent",
			err:  ErrPaymasterRejected,
			want: true,
		},
		{
			give: "ErrInsufficientToken is permanent",
			err:  ErrInsufficientToken,
			want: true,
		},
		{
			give: "wrapped ErrPaymasterRejected is permanent",
			err:  fmt.Errorf("check: %w", ErrPaymasterRejected),
			want: true,
		},
		{
			give: "wrapped ErrInsufficientToken is permanent",
			err:  fmt.Errorf("balance: %w", ErrInsufficientToken),
			want: true,
		},
		{
			give: "double-wrapped permanent error",
			err:  fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrPaymasterRejected)),
			want: true,
		},
		{
			give: "ErrPaymasterTimeout is not permanent",
			err:  ErrPaymasterTimeout,
			want: false,
		},
		{
			give: "ErrPaymasterNotConfigured is not permanent",
			err:  ErrPaymasterNotConfigured,
			want: false,
		},
		{
			give: "unrelated error is not permanent",
			err:  errors.New("random failure"),
			want: false,
		},
		{
			give: "nil error is not permanent",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsPermanent(tt.err))
		})
	}
}

func TestTransientAndPermanent_MutuallyExclusive(t *testing.T) {
	t.Parallel()

	// Every sentinel error should be at most one of transient or permanent.
	sentinels := []struct {
		give string
		err  error
	}{
		{give: "ErrPaymasterRejected", err: ErrPaymasterRejected},
		{give: "ErrPaymasterTimeout", err: ErrPaymasterTimeout},
		{give: "ErrInsufficientToken", err: ErrInsufficientToken},
		{give: "ErrPaymasterNotConfigured", err: ErrPaymasterNotConfigured},
	}

	for _, tt := range sentinels {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			transient := IsTransient(tt.err)
			permanent := IsPermanent(tt.err)
			assert.False(t, transient && permanent,
				"%s should not be both transient and permanent", tt.give)
		})
	}
}

func TestSentinelErrors_DistinctMessages(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		ErrPaymasterRejected,
		ErrPaymasterTimeout,
		ErrInsufficientToken,
		ErrPaymasterNotConfigured,
	}

	seen := make(map[string]bool, len(sentinels))
	for _, err := range sentinels {
		msg := err.Error()
		assert.False(t, seen[msg], "duplicate error message: %s", msg)
		seen[msg] = true
	}
}
