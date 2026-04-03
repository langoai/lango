package agentrt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ctxkeys"
)

func TestNewRecursionGuard_DefaultDepth(t *testing.T) {
	guard := NewRecursionGuard(0)
	assert.Equal(t, 3, guard.MaxDepth)

	guard2 := NewRecursionGuard(-1)
	assert.Equal(t, 3, guard2.MaxDepth)
}

func TestNewRecursionGuard_CustomDepth(t *testing.T) {
	guard := NewRecursionGuard(5)
	assert.Equal(t, 5, guard.MaxDepth)
}

func TestRecursionGuard_NormalAllowed(t *testing.T) {
	guard := NewRecursionGuard(3)
	ctx := ctxkeys.WithSpawnDepth(context.Background(), 0)

	err := guard.Check(ctx, "orchestrator", "worker")
	require.NoError(t, err)
}

func TestRecursionGuard_DepthExceeded(t *testing.T) {
	tests := []struct {
		give      string
		giveDepth int
		giveMax   int
	}{
		{
			give:      "at max depth",
			giveDepth: 3,
			giveMax:   3,
		},
		{
			give:      "above max depth",
			giveDepth: 5,
			giveMax:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			guard := NewRecursionGuard(tt.giveMax)
			ctx := ctxkeys.WithSpawnDepth(context.Background(), tt.giveDepth)

			err := guard.Check(ctx, "parent", "child")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "spawn depth")
			assert.Contains(t, err.Error(), "exceeds max")
		})
	}
}

func TestRecursionGuard_SelfSpawnBlocked(t *testing.T) {
	guard := NewRecursionGuard(3)
	ctx := ctxkeys.WithSpawnDepth(context.Background(), 0)

	err := guard.Check(ctx, "worker", "worker")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "self-spawn blocked")
	assert.Contains(t, err.Error(), "worker")
}

func TestRecursionGuard_SelfSpawnAllowedWhenSpawnerEmpty(t *testing.T) {
	guard := NewRecursionGuard(3)
	ctx := ctxkeys.WithSpawnDepth(context.Background(), 0)

	// Empty spawner should not trigger self-spawn check.
	err := guard.Check(ctx, "", "worker")
	require.NoError(t, err)
}

func TestRecursionGuard_CycleDetected(t *testing.T) {
	guard := NewRecursionGuard(10)
	ctx := ctxkeys.WithSpawnDepth(context.Background(), 1)
	ctx = ctxkeys.WithSpawnChain(ctx, []string{"orchestrator", "planner"})

	err := guard.Check(ctx, "planner", "orchestrator")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
	assert.Contains(t, err.Error(), "orchestrator")
}

func TestRecursionGuard_NoCycleWhenChainEmpty(t *testing.T) {
	guard := NewRecursionGuard(10)
	ctx := ctxkeys.WithSpawnDepth(context.Background(), 0)

	err := guard.Check(ctx, "parent", "child")
	require.NoError(t, err)
}

func TestRecursionGuard_NoCycleWhenTargetNotInChain(t *testing.T) {
	guard := NewRecursionGuard(10)
	ctx := ctxkeys.WithSpawnDepth(context.Background(), 2)
	ctx = ctxkeys.WithSpawnChain(ctx, []string{"orchestrator", "planner"})

	err := guard.Check(ctx, "planner", "worker")
	require.NoError(t, err)
}

func TestRecursionGuard_DepthCheckBeforeSelfSpawn(t *testing.T) {
	// When depth is exceeded, the depth error should be returned
	// even if self-spawn would also be blocked.
	guard := NewRecursionGuard(2)
	ctx := ctxkeys.WithSpawnDepth(context.Background(), 3)

	err := guard.Check(ctx, "worker", "worker")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "spawn depth")
}

func TestRecursionGuard_EmptyContext(t *testing.T) {
	guard := NewRecursionGuard(3)

	// Empty context has depth 0, no chain — should pass.
	err := guard.Check(context.Background(), "parent", "child")
	require.NoError(t, err)
}
