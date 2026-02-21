package learning

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	entlearning "github.com/langowarny/lango/internal/ent/learning"
)

func TestExtractErrorPattern(t *testing.T) {
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
			got := extractErrorPattern(errors.New(tt.give))
			if got != tt.want {
				t.Errorf("extractErrorPattern(%q) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}

func TestCategorizeError(t *testing.T) {
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
			got := categorizeError(tt.giveTool, tt.giveErr)
			if got != tt.want {
				t.Errorf("categorizeError(%q, %v) = %q, want %q", tt.giveTool, tt.giveErr, got, tt.want)
			}
		})
	}
}

func TestIsDeadlineExceeded(t *testing.T) {
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
			got := isDeadlineExceeded(tt.err)
			if got != tt.want {
				t.Errorf("isDeadlineExceeded(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestSummarizeParams(t *testing.T) {
	longStr := strings.Repeat("a", 250)

	t.Run("nil params returns nil", func(t *testing.T) {
		got := summarizeParams(nil)
		if got != nil {
			t.Fatalf("summarizeParams(nil) = %v, want nil", got)
		}
	})

	t.Run("short string stays unchanged", func(t *testing.T) {
		give := map[string]interface{}{"key": "hello"}
		got := summarizeParams(give)
		if val, ok := got["key"]; !ok || val != "hello" {
			t.Errorf("summarizeParams short string: got %v, want %q", got["key"], "hello")
		}
	})

	t.Run("long string truncated to 203 chars", func(t *testing.T) {
		give := map[string]interface{}{"key": longStr}
		got := summarizeParams(give)
		val, ok := got["key"].(string)
		if !ok {
			t.Fatalf("expected string, got %T", got["key"])
		}
		if len(val) != 203 {
			t.Errorf("truncated length = %d, want 203", len(val))
		}
		if !strings.HasSuffix(val, "...") {
			t.Errorf("truncated string should end with '...', got suffix %q", val[len(val)-3:])
		}
	})

	t.Run("slice becomes [N items]", func(t *testing.T) {
		give := map[string]interface{}{
			"list": []interface{}{1, 2, 3},
		}
		got := summarizeParams(give)
		if val, ok := got["list"]; !ok || val != "[3 items]" {
			t.Errorf("summarizeParams slice: got %v, want %q", got["list"], "[3 items]")
		}
	})

	t.Run("int stays unchanged", func(t *testing.T) {
		give := map[string]interface{}{"count": 42}
		got := summarizeParams(give)
		if val, ok := got["count"]; !ok || val != 42 {
			t.Errorf("summarizeParams int: got %v, want %d", got["count"], 42)
		}
	})
}
