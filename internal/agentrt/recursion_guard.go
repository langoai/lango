package agentrt

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/ctxkeys"
)

// RecursionGuard prevents runaway agent spawn recursion by enforcing
// depth limits, self-spawn rejection, and cycle detection.
type RecursionGuard struct {
	MaxDepth int
}

// NewRecursionGuard creates a RecursionGuard with the given maximum spawn depth.
// If maxDepth <= 0, the default of 3 is used.
func NewRecursionGuard(maxDepth int) *RecursionGuard {
	if maxDepth <= 0 {
		maxDepth = 3
	}
	return &RecursionGuard{MaxDepth: maxDepth}
}

// Check verifies that a spawn from spawner to target is allowed under the
// current context. It returns an error if:
//   - SpawnDepth from context >= MaxDepth (depth exceeded)
//   - spawner == target (self-spawn)
//   - target already appears in the spawn chain (cycle detected)
func (g *RecursionGuard) Check(ctx context.Context, spawner, target string) error {
	depth := ctxkeys.SpawnDepthFromContext(ctx)
	if depth >= g.MaxDepth {
		return fmt.Errorf("recursion guard: spawn depth %d exceeds max %d", depth, g.MaxDepth)
	}

	if spawner != "" && spawner == target {
		return fmt.Errorf("recursion guard: self-spawn blocked (%q cannot spawn itself)", spawner)
	}

	chain := ctxkeys.SpawnChainFromContext(ctx)
	for _, name := range chain {
		if name == target {
			return fmt.Errorf("recursion guard: cycle detected (%q already in spawn chain %v)", target, chain)
		}
	}

	return nil
}
