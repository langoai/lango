package mcp

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrServerNotFound", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, ErrServerNotFound, "mcp: server not found")
	})

	t.Run("ErrConnectionFailed", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, ErrConnectionFailed, "mcp: connection failed")
	})

	t.Run("ErrToolCallFailed", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, ErrToolCallFailed, "mcp: tool call failed")
	})

	t.Run("ErrNotConnected", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, ErrNotConnected, "mcp: not connected")
	})

	t.Run("ErrInvalidTransport", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, ErrInvalidTransport, "mcp: invalid transport type")
	})
}

func TestSentinelErrors_Wrapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give   string
		target error
	}{
		{give: "ErrServerNotFound", target: ErrServerNotFound},
		{give: "ErrConnectionFailed", target: ErrConnectionFailed},
		{give: "ErrToolCallFailed", target: ErrToolCallFailed},
		{give: "ErrNotConnected", target: ErrNotConnected},
		{give: "ErrInvalidTransport", target: ErrInvalidTransport},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			wrapped := fmt.Errorf("context: %w", tt.target)
			assert.True(t, errors.Is(wrapped, tt.target))
		})
	}
}

func TestSentinelErrors_AreDistinct(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		ErrServerNotFound,
		ErrConnectionFailed,
		ErrToolCallFailed,
		ErrNotConnected,
		ErrInvalidTransport,
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j {
				assert.NotErrorIs(t, a, b, "sentinel %d should not match sentinel %d", i, j)
			}
		}
	}
}
