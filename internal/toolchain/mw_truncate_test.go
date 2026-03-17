package toolchain

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTruncate(t *testing.T) {
	tests := []struct {
		give       string
		maxChars   int
		result     interface{}
		wantErr    error
		wantResult interface{}
		wantTrunc  bool
	}{
		{
			give:       "under limit string",
			maxChars:   100,
			result:     "short text",
			wantResult: "short text",
		},
		{
			give:       "over limit string",
			maxChars:   10,
			result:     "this is a very long string that exceeds the limit",
			wantTrunc:  true,
		},
		{
			give:       "map result over limit",
			maxChars:   10,
			result:     map[string]string{"key": "a long value that should be truncated"},
			wantTrunc:  true,
		},
		{
			give:       "error passes through",
			maxChars:   10,
			result:     "some result",
			wantErr:    errors.New("tool failed"),
			wantResult: "some result",
		},
		{
			give:       "zero maxChars uses default",
			maxChars:   0,
			result:     "short text",
			wantResult: "short text",
		},
		{
			give:       "negative maxChars uses default",
			maxChars:   -5,
			result:     "short text",
			wantResult: "short text",
		},
		{
			give:       "nil result passes through",
			maxChars:   10,
			result:     nil,
			wantResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			tool := &agent.Tool{Name: "test_tool"}
			handler := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				return tt.result, tt.wantErr
			}

			mw := WithTruncate(tt.maxChars)
			wrapped := mw(tool, handler)

			got, err := wrapped(context.Background(), nil)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantResult, got)
				return
			}

			require.NoError(t, err)

			if tt.wantTrunc {
				s, ok := got.(string)
				require.True(t, ok, "truncated result should be a string")
				assert.True(t, strings.HasSuffix(s, "\n... [output truncated]"))
				// The truncated content should be maxChars + marker length
				assert.Equal(t, tt.maxChars, len(s)-len("\n... [output truncated]"))
			} else {
				assert.Equal(t, tt.wantResult, got)
			}
		})
	}
}

func TestWithTruncateDefaultMaxChars(t *testing.T) {
	// Verify that a string just at the default limit passes through.
	tool := &agent.Tool{Name: "test_tool"}
	text := strings.Repeat("x", defaultMaxOutputChars)
	handler := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		return text, nil
	}

	mw := WithTruncate(0)
	wrapped := mw(tool, handler)
	got, err := wrapped(context.Background(), nil)

	require.NoError(t, err)
	assert.Equal(t, text, got)

	// One char over the default should truncate.
	text2 := strings.Repeat("x", defaultMaxOutputChars+1)
	handler2 := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		return text2, nil
	}
	wrapped2 := mw(tool, handler2)
	got2, err := wrapped2(context.Background(), nil)

	require.NoError(t, err)
	s, ok := got2.(string)
	require.True(t, ok)
	assert.True(t, strings.HasSuffix(s, "\n... [output truncated]"))
}
