package adk

import (
	"context"

	"github.com/langoai/lango/internal/ctxkeys"
)

// WithAgentName returns a context carrying the given agent name.
// This delegates to ctxkeys.WithAgentName so that any package importing
// ctxkeys can read the value without depending on the adk package.
func WithAgentName(ctx context.Context, name string) context.Context {
	return ctxkeys.WithAgentName(ctx, name)
}

// AgentNameFromContext extracts the agent name stored in ctx.
// Returns an empty string when no agent name is present.
func AgentNameFromContext(ctx context.Context) string {
	return ctxkeys.AgentNameFromContext(ctx)
}
