package toolchain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

// TestMiddlewareChain_ProductionOrder reconstructs the same middleware chain
// as app.go Phase B4 (steps B4a–B4e) and verifies the invocation order.
//
// Production order (outermost → innermost):
//   ExecPolicy → Approval → Principal → Hooks → OutputManager → Learning → Handler
//
// This is built by successive ChainAll calls in app.go:
//   B4a: WithLearning       (innermost — applied first)
//   B4b: WithOutputManager
//   B4c: WithHooks
//   B4c2: WithPrincipal
//   B4d: WithApproval
//   B4e: WithPolicy         (outermost — applied last)
func TestMiddlewareChain_ProductionOrder(t *testing.T) {
	t.Parallel()

	var order []string

	labelMiddleware := func(label string) Middleware {
		return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
			return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				order = append(order, label+":enter")
				result, err := next(ctx, params)
				order = append(order, label+":exit")
				return result, err
			}
		}
	}

	tool := &agent.Tool{
		Name:        "exec",
		SafetyLevel: agent.SafetyLevelDangerous,
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			order = append(order, "handler")
			return "ok", nil
		},
	}

	// Apply middlewares in the same order as app.go Phase B4.
	// Each ChainAll call wraps tools with a new outermost layer.
	tools := []*agent.Tool{tool}
	tools = ChainAll(tools, labelMiddleware("learning"))       // B4a — innermost
	tools = ChainAll(tools, labelMiddleware("output_manager")) // B4b
	tools = ChainAll(tools, labelMiddleware("hooks"))          // B4c
	tools = ChainAll(tools, labelMiddleware("principal"))      // B4c2
	tools = ChainAll(tools, labelMiddleware("approval"))       // B4d
	tools = ChainAll(tools, labelMiddleware("exec_policy"))    // B4e — outermost

	wrapped := tools[0]
	_, err := wrapped.Handler(context.Background(), nil)
	require.NoError(t, err)

	// Expected invocation order: outermost enters first, innermost exits first.
	wantOrder := []string{
		"exec_policy:enter",
		"approval:enter",
		"principal:enter",
		"hooks:enter",
		"output_manager:enter",
		"learning:enter",
		"handler",
		"learning:exit",
		"output_manager:exit",
		"hooks:exit",
		"principal:exit",
		"approval:exit",
		"exec_policy:exit",
	}
	assert.Equal(t, wantOrder, order, "middleware chain invocation order must match production order")
}

// TestMiddlewareChain_PolicyBlocksBeforeApproval verifies that when ExecPolicy
// blocks a command, the Approval middleware (and all inner layers) are never reached.
// This ensures users are never asked to approve a command that will be blocked anyway.
func TestMiddlewareChain_PolicyBlocksBeforeApproval(t *testing.T) {
	t.Parallel()

	var reached []string

	trackMiddleware := func(label string) Middleware {
		return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
			return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				reached = append(reached, label)
				return next(ctx, params)
			}
		}
	}

	blockingPolicy := func(_ *agent.Tool, _ agent.ToolHandler) agent.ToolHandler {
		return func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			reached = append(reached, "exec_policy")
			return map[string]interface{}{"blocked": true, "message": "dangerous command blocked"}, nil
		}
	}

	tool := &agent.Tool{
		Name:        "exec",
		SafetyLevel: agent.SafetyLevelDangerous,
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			reached = append(reached, "handler")
			return "executed", nil
		},
	}

	tools := []*agent.Tool{tool}
	tools = ChainAll(tools, trackMiddleware("learning"))
	tools = ChainAll(tools, trackMiddleware("output_manager"))
	tools = ChainAll(tools, trackMiddleware("hooks"))
	tools = ChainAll(tools, trackMiddleware("principal"))
	tools = ChainAll(tools, trackMiddleware("approval"))
	tools = ChainAll(tools, blockingPolicy) // outermost — blocks before approval

	wrapped := tools[0]
	result, err := wrapped.Handler(context.Background(), map[string]interface{}{"command": "kill 1234"})
	require.NoError(t, err)

	// Only ExecPolicy should have been reached.
	assert.Equal(t, []string{"exec_policy"}, reached,
		"when ExecPolicy blocks, no inner middleware should be reached")

	// Verify the blocked result structure.
	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, m["blocked"])
}

// TestMiddlewareChain_ApprovalDenialBlocksInnerLayers verifies that when approval
// is denied, inner layers (hooks, output manager, learning, handler) are not reached.
func TestMiddlewareChain_ApprovalDenialBlocksInnerLayers(t *testing.T) {
	t.Parallel()

	var reached []string

	trackMiddleware := func(label string) Middleware {
		return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
			return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				reached = append(reached, label)
				return next(ctx, params)
			}
		}
	}

	denyingApproval := func(_ *agent.Tool, _ agent.ToolHandler) agent.ToolHandler {
		return func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			reached = append(reached, "approval")
			return nil, assert.AnError
		}
	}

	tool := &agent.Tool{
		Name:        "exec",
		SafetyLevel: agent.SafetyLevelDangerous,
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			reached = append(reached, "handler")
			return "executed", nil
		},
	}

	tools := []*agent.Tool{tool}
	tools = ChainAll(tools, trackMiddleware("learning"))
	tools = ChainAll(tools, trackMiddleware("output_manager"))
	tools = ChainAll(tools, trackMiddleware("hooks"))
	tools = ChainAll(tools, trackMiddleware("principal"))
	tools = ChainAll(tools, denyingApproval)
	tools = ChainAll(tools, trackMiddleware("exec_policy"))

	wrapped := tools[0]
	_, err := wrapped.Handler(context.Background(), nil)
	require.Error(t, err)

	// ExecPolicy should pass through, then approval should deny.
	assert.Equal(t, []string{"exec_policy", "approval"}, reached,
		"when approval denies, inner layers should not be reached")
}

// TestMiddlewareChain_MultipleToolsShareOrder verifies that ChainAll applies
// the same middleware stack to all tools consistently.
func TestMiddlewareChain_MultipleToolsShareOrder(t *testing.T) {
	t.Parallel()

	type toolExec struct {
		name  string
		order []string
	}

	toolNames := []string{"exec", "fs_write", "browser_navigate"}
	executions := make([]toolExec, len(toolNames))

	baseTools := make([]*agent.Tool, len(toolNames))
	for i, name := range toolNames {
		baseTools[i] = &agent.Tool{
			Name:        name,
			SafetyLevel: agent.SafetyLevelDangerous,
			Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
				executions[i].order = append(executions[i].order, "handler")
				return "ok", nil
			},
		}
		executions[i].name = name
	}

	labelMiddleware := func(label string) Middleware {
		return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
			return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				// Find the execution by matching the tool name.
				for i := range executions {
					if executions[i].name == tool.Name {
						executions[i].order = append(executions[i].order, label)
						break
					}
				}
				return next(ctx, params)
			}
		}
	}

	tools := baseTools
	tools = ChainAll(tools, labelMiddleware("learning"))
	tools = ChainAll(tools, labelMiddleware("output_manager"))
	tools = ChainAll(tools, labelMiddleware("hooks"))
	tools = ChainAll(tools, labelMiddleware("principal"))
	tools = ChainAll(tools, labelMiddleware("approval"))
	tools = ChainAll(tools, labelMiddleware("exec_policy"))

	wantOrder := []string{"exec_policy", "approval", "principal", "hooks", "output_manager", "learning", "handler"}

	for i, tool := range tools {
		_, err := tool.Handler(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, wantOrder, executions[i].order,
			"tool %q should have same middleware order", executions[i].name)
	}
}
