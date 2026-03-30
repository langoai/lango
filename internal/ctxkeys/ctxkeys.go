// Package ctxkeys provides shared context keys for cross-package value propagation.
// It exists as a lightweight, dependency-free package so that both adk and toolchain
// (and any future packages) can read/write the same context values without import cycles.
package ctxkeys

import "context"

type contextKey string

const (
	agentNameKey contextKey = "lango.agent_name"
	principalKey contextKey = "lango.principal"
)

// WithAgentName returns a new context carrying the given agent name.
func WithAgentName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, agentNameKey, name)
}

// AgentNameFromContext extracts the agent name from ctx.
// It returns an empty string if no agent name is present.
func AgentNameFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(agentNameKey).(string); ok {
		return v
	}
	return ""
}

// WithPrincipal returns a new context carrying the given principal name.
func WithPrincipal(ctx context.Context, principal string) context.Context {
	return context.WithValue(ctx, principalKey, principal)
}

// PrincipalFromContext extracts the principal from ctx.
// It returns an empty string if no principal is present.
func PrincipalFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(principalKey).(string); ok {
		return v
	}
	return ""
}
