package ctxkeys

import (
	"context"
	"testing"
)

func TestAgentNameRoundtrip(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{give: "planner", want: "planner"},
		{give: "executor", want: "executor"},
		{give: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			ctx := WithAgentName(context.Background(), tt.give)
			got := AgentNameFromContext(ctx)
			if got != tt.want {
				t.Errorf("AgentNameFromContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAgentNameFromContext_EmptyContext(t *testing.T) {
	got := AgentNameFromContext(context.Background())
	if got != "" {
		t.Errorf("AgentNameFromContext(empty) = %q, want empty string", got)
	}
}

func TestAgentNameOverwrite(t *testing.T) {
	ctx := WithAgentName(context.Background(), "first")
	ctx = WithAgentName(ctx, "second")

	got := AgentNameFromContext(ctx)
	if got != "second" {
		t.Errorf("AgentNameFromContext() = %q, want %q", got, "second")
	}
}
