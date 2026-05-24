package toolchain

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
)

// ModeAllowlistResolver resolves the active session mode's tool allowlist
// (set of tool names) for a given context. Returns an empty map and false
// when no mode is active or the mode has no tool allowlist (no enforcement).
type ModeAllowlistResolver func(ctx context.Context) (allow map[string]bool, active bool)

// WithModeAllowlist returns a Middleware that blocks tool execution when the
// current session has an active mode and the tool name is not in the mode's
// allowlist. When no mode is active, the middleware passes through.
//
// Enforcement at the middleware level (not the dispatcher) is required because
// the agent receives the full tool set at boot via adk.NewAgent and dispatches
// directly to handlers — direct calls bypass any dispatcher-level filter.
//
// The error message identifies both the blocked tool and the active mode so
// the LLM can recover on the next turn.
func WithModeAllowlist(resolver ModeAllowlistResolver) Middleware {
	return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		return func(ctx context.Context, params map[string]any) (any, error) {
			if resolver == nil {
				return next(ctx, params)
			}
			allow, active := resolver(ctx)
			if !active {
				return next(ctx, params)
			}
			if allow[tool.Name] {
				return next(ctx, params)
			}
			return nil, fmt.Errorf("tool %q not available in current mode", tool.Name)
		}
	}
}
