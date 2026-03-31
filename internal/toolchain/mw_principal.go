package toolchain

import (
	"context"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
)

// WithPrincipal returns a Middleware that copies the agent name from context
// into the principal context key. This bridges the ADK agent-name injection
// (adk/tools.go) with the ontology ACL layer (ontology/service.go).
//
// Injection point: B4c2 in the middleware chain (after WithHooks, before WithApproval).
// Programmatic callers (SeedDefaults, internal wiring) bypass this middleware,
// so PrincipalFromContext returns "" for them — treated as "system" by ACL.
func WithPrincipal() Middleware {
	return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		return func(ctx context.Context, params map[string]any) (any, error) {
			if name := ctxkeys.AgentNameFromContext(ctx); name != "" {
				ctx = ctxkeys.WithPrincipal(ctx, name)
			}
			return next(ctx, params)
		}
	}
}
