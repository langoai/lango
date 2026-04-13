package exec

import (
	"context"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/toolchain"
)

// WithPolicy returns a Middleware that evaluates exec/exec_bg commands against
// the PolicyEvaluator before any other middleware (including approval).
// Block verdicts return BlockedResult without calling next.
// Observe verdicts log/publish then call next.
// Allow verdicts and non-exec tools pass through unchanged.
func WithPolicy(pe *PolicyEvaluator) toolchain.Middleware {
	return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Only evaluate exec and exec_bg tools.
			if tool.Name != "exec" && tool.Name != "exec_bg" {
				return next(ctx, params)
			}

			cmd, ok := params["command"].(string)
			if !ok || cmd == "" {
				return next(ctx, params)
			}

			d := pe.Evaluate(cmd)
			switch d.Verdict {
			case VerdictBlock:
				pe.publishAndLog(d, ctx)
				return &BlockedResult{Blocked: true, Message: d.Message}, nil
			case VerdictObserve:
				pe.publishAndLog(d, ctx)
				return next(ctx, params)
			default:
				return next(ctx, params)
			}
		}
	}
}
