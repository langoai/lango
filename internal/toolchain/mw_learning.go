package toolchain

import (
	"context"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/learning"
	"github.com/langoai/lango/internal/session"
)

// WithLearning returns a middleware that observes tool results for learning.
// After each handler execution the observer is called with session key, tool name,
// parameters, result, and any error.
func WithLearning(observer learning.ToolResultObserver) Middleware {
	return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			result, err := next(ctx, params)
			sessionKey := session.SessionKeyFromContext(ctx)
			observer.OnToolResult(ctx, sessionKey, tool.Name, params, result, err)
			return result, err
		}
	}
}
