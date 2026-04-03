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

func TestDynamicAllowedToolsRoundtrip(t *testing.T) {
	tests := []struct {
		give []string
		want []string
	}{
		{give: []string{"fs_read", "web_search"}, want: []string{"fs_read", "web_search"}},
		{give: []string{}, want: []string{}},
		{give: nil, want: nil},
	}

	for _, tt := range tests {
		ctx := context.Background()
		if tt.give != nil {
			ctx = WithDynamicAllowedTools(ctx, tt.give)
		}
		got := DynamicAllowedToolsFromContext(ctx)
		if tt.want == nil {
			if got != nil {
				t.Errorf("DynamicAllowedToolsFromContext() = %v, want nil", got)
			}
		} else {
			if len(got) != len(tt.want) {
				t.Errorf("DynamicAllowedToolsFromContext() len = %d, want %d", len(got), len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("DynamicAllowedToolsFromContext()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		}
	}
}

func TestDynamicAllowedToolsFromContext_EmptyContext(t *testing.T) {
	got := DynamicAllowedToolsFromContext(context.Background())
	if got != nil {
		t.Errorf("DynamicAllowedToolsFromContext(empty) = %v, want nil", got)
	}
}

func TestSpawnDepthRoundtrip(t *testing.T) {
	tests := []struct {
		give int
		want int
	}{
		{give: 0, want: 0},
		{give: 1, want: 1},
		{give: 5, want: 5},
	}

	for _, tt := range tests {
		ctx := WithSpawnDepth(context.Background(), tt.give)
		got := SpawnDepthFromContext(ctx)
		if got != tt.want {
			t.Errorf("SpawnDepthFromContext() = %d, want %d", got, tt.want)
		}
	}
}

func TestSpawnDepthFromContext_EmptyContext(t *testing.T) {
	got := SpawnDepthFromContext(context.Background())
	if got != 0 {
		t.Errorf("SpawnDepthFromContext(empty) = %d, want 0", got)
	}
}

func TestSpawnDepthOverwrite(t *testing.T) {
	ctx := WithSpawnDepth(context.Background(), 3)
	ctx = WithSpawnDepth(ctx, 7)

	got := SpawnDepthFromContext(ctx)
	if got != 7 {
		t.Errorf("SpawnDepthFromContext() = %d, want %d", got, 7)
	}
}
