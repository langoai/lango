package learning

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entlearning "github.com/langoai/lango/internal/ent/learning"
)

func TestExtractErrorPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want string
	}{
		{
			give: "error with uuid a1b2c3d4-e5f6-7890-abcd-ef1234567890 inside",
			want: "error with uuid  inside",
		},
		{
			give: "failed at 2024-01-15T10:30:00 during sync",
			want: "failed at  during sync",
		},
		{
			give: "failed at 2024-01-15 10:30:00 during sync",
			want: "failed at  during sync",
		},
		{
			give: "error reading /home/user/data/ config",
			want: "error reading <path> config",
		},
		{
			give: "connection to server:8080 refused",
			want: "connection to server:<port> refused",
		},
		{
			give: "uuid a1b2c3d4-e5f6-7890-abcd-ef1234567890 at 2024-01-15T10:30:00 path /var/log/app/ on port:9090",
			want: "uuid  at  path <path> on port:<port>",
		},
		{
			give: "simple error message",
			want: "simple error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := extractErrorPattern(errors.New(tt.give))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCategorizeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		giveErr  error
		giveTool string
		want     entlearning.Category
	}{
		{
			give:     "context.DeadlineExceeded",
			giveErr:  context.DeadlineExceeded,
			giveTool: "",
			want:     entlearning.CategoryTimeout,
		},
		{
			give:     "deadline exceeded string",
			giveErr:  errors.New("deadline exceeded waiting for response"),
			giveTool: "",
			want:     entlearning.CategoryTimeout,
		},
		{
			give:     "timeout string",
			giveErr:  errors.New("connection timeout"),
			giveTool: "",
			want:     entlearning.CategoryTimeout,
		},
		{
			give:     "permission denied",
			giveErr:  errors.New("permission denied"),
			giveTool: "",
			want:     entlearning.CategoryPermission,
		},
		{
			give:     "access denied",
			giveErr:  errors.New("access denied for user"),
			giveTool: "",
			want:     entlearning.CategoryPermission,
		},
		{
			give:     "forbidden",
			giveErr:  errors.New("forbidden resource"),
			giveTool: "",
			want:     entlearning.CategoryPermission,
		},
		{
			give:     "api error",
			giveErr:  errors.New("api call failed"),
			giveTool: "",
			want:     entlearning.CategoryProviderError,
		},
		{
			give:     "model error",
			giveErr:  errors.New("model not found"),
			giveTool: "",
			want:     entlearning.CategoryProviderError,
		},
		{
			give:     "provider error",
			giveErr:  errors.New("provider unavailable"),
			giveTool: "",
			want:     entlearning.CategoryProviderError,
		},
		{
			give:     "rate limit",
			giveErr:  errors.New("rate limit exceeded"),
			giveTool: "",
			want:     entlearning.CategoryProviderError,
		},
		{
			give:     "tool error with toolName",
			giveErr:  errors.New("something broke"),
			giveTool: "exec",
			want:     entlearning.CategoryToolError,
		},
		{
			give:     "general error no toolName",
			giveErr:  errors.New("something broke"),
			giveTool: "",
			want:     entlearning.CategoryGeneral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := categorizeError(tt.giveTool, tt.giveErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsDeadlineExceeded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		err  error
		want bool
	}{
		{
			give: "direct DeadlineExceeded",
			err:  context.DeadlineExceeded,
			want: true,
		},
		{
			give: "wrapped DeadlineExceeded",
			err:  fmt.Errorf("outer: %w", context.DeadlineExceeded),
			want: true,
		},
		{
			give: "regular error",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := isDeadlineExceeded(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSummarizeParams(t *testing.T) {
	t.Parallel()

	longStr := strings.Repeat("a", 250)

	t.Run("nil params returns nil", func(t *testing.T) {
		t.Parallel()
		got := summarizeParams(nil)
		assert.Nil(t, got)
	})

	t.Run("short string stays unchanged", func(t *testing.T) {
		t.Parallel()
		give := map[string]interface{}{"key": "hello"}
		got := summarizeParams(give)
		assert.Equal(t, "hello", got["key"])
	})

	t.Run("long string truncated to 203 chars", func(t *testing.T) {
		t.Parallel()
		give := map[string]interface{}{"key": longStr}
		got := summarizeParams(give)
		val, ok := got["key"].(string)
		require.True(t, ok, "expected string, got %T", got["key"])
		assert.Len(t, val, 203)
		assert.True(t, strings.HasSuffix(val, "..."), "truncated string should end with '...'")
	})

	t.Run("slice becomes [N items]", func(t *testing.T) {
		t.Parallel()
		give := map[string]interface{}{
			"list": []interface{}{1, 2, 3},
		}
		got := summarizeParams(give)
		assert.Equal(t, "[3 items]", got["list"])
	})

	t.Run("int stays unchanged", func(t *testing.T) {
		t.Parallel()
		give := map[string]interface{}{"count": 42}
		got := summarizeParams(give)
		assert.Equal(t, 42, got["count"])
	})
}
