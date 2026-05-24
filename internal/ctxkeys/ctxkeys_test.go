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

func TestPrincipalRoundtripAndOverwrite(t *testing.T) {
	ctx := WithPrincipal(context.Background(), "operator:alice")
	if got := PrincipalFromContext(ctx); got != "operator:alice" {
		t.Fatalf("PrincipalFromContext() = %q, want %q", got, "operator:alice")
	}

	ctx = WithPrincipal(ctx, "agent:planner")
	if got := PrincipalFromContext(ctx); got != "agent:planner" {
		t.Fatalf("PrincipalFromContext() after overwrite = %q, want %q", got, "agent:planner")
	}
}

func TestPrincipalFromContextEmpty(t *testing.T) {
	if got := PrincipalFromContext(context.Background()); got != "" {
		t.Fatalf("PrincipalFromContext(empty) = %q, want empty string", got)
	}
}

func TestP2PRequestMarker(t *testing.T) {
	if IsP2PRequest(context.Background()) {
		t.Fatal("IsP2PRequest(empty) = true, want false")
	}

	ctx := WithP2PRequest(context.Background())
	if !IsP2PRequest(ctx) {
		t.Fatal("IsP2PRequest(marked) = false, want true")
	}
}

func TestMissionIDRoundtripAndEmpty(t *testing.T) {
	if got := MissionIDFromContext(context.Background()); got != "" {
		t.Fatalf("MissionIDFromContext(empty) = %q, want empty string", got)
	}

	ctx := WithMissionID(context.Background(), "mission-123")
	if got := MissionIDFromContext(ctx); got != "mission-123" {
		t.Fatalf("MissionIDFromContext() = %q, want %q", got, "mission-123")
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

func TestSpawnChainRoundtripAndEmpty(t *testing.T) {
	if got := SpawnChainFromContext(context.Background()); got != nil {
		t.Fatalf("SpawnChainFromContext(empty) = %v, want nil", got)
	}

	want := []string{"root", "child"}
	ctx := WithSpawnChain(context.Background(), want)
	got := SpawnChainFromContext(ctx)
	if len(got) != len(want) {
		t.Fatalf("SpawnChainFromContext() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SpawnChainFromContext()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestWithDefaultPrincipalPreservesExistingPrincipal(t *testing.T) {
	ctx := WithPrincipal(context.Background(), "explicit")
	got := WithDefaultPrincipal(ctx, "fallback")

	if principal := PrincipalFromContext(got); principal != "explicit" {
		t.Fatalf("PrincipalFromContext() = %q, want %q", principal, "explicit")
	}
}

func TestWithDefaultPrincipalPrefersAgentName(t *testing.T) {
	ctx := WithAgentName(context.Background(), "agent:planner")
	got := WithDefaultPrincipal(ctx, "fallback")

	if principal := PrincipalFromContext(got); principal != "agent:planner" {
		t.Fatalf("PrincipalFromContext() = %q, want %q", principal, "agent:planner")
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
