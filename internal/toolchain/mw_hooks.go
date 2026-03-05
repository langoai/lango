package toolchain

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/session"
)

// WithHooks returns a Middleware that integrates the HookRegistry into the
// existing middleware chain. Flow: RunPre -> (if Continue/Modify) next(params) -> RunPost.
func WithHooks(registry *HookRegistry) Middleware {
	return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			hctx := HookContext{
				ToolName:   tool.Name,
				AgentName:  AgentNameFromContext(ctx),
				Params:     params,
				SessionKey: session.SessionKeyFromContext(ctx),
				Ctx:        ctx,
			}

			// Run pre-hooks.
			preResult, err := registry.RunPre(hctx)
			if err != nil {
				return nil, fmt.Errorf("pre-hook %s: %w", tool.Name, err)
			}

			switch preResult.Action {
			case Block:
				return nil, fmt.Errorf("tool '%s' blocked by hook: %s", tool.Name, preResult.BlockReason)
			case Modify:
				params = preResult.ModifiedParams
			}

			// Execute the tool.
			result, toolErr := next(ctx, params)

			// Run post-hooks (errors are logged, not propagated).
			postErr := registry.RunPost(hctx, result, toolErr)
			if postErr != nil {
				logging.App().Warnw("post-hook error", "tool", tool.Name, "error", postErr)
			}

			return result, toolErr
		}
	}
}
