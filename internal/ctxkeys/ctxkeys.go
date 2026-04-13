// Package ctxkeys provides shared context keys for cross-package value propagation.
// It exists as a lightweight, dependency-free package so that both adk and toolchain
// (and any future packages) can read/write the same context values without import cycles.
package ctxkeys

import "context"

type contextKey string

const (
	agentNameKey         contextKey = "lango.agent_name"
	principalKey         contextKey = "lango.principal"
	p2pRequestKey        contextKey = "lango.p2p_request"
	dynamicAllowedTools  contextKey = "lango.dynamic_allowed_tools"
	spawnDepthKey        contextKey = "lango.spawn_depth"
	spawnChainKey        contextKey = "lango.spawn_chain"
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

// WithP2PRequest returns a new context marked as originating from a P2P peer request.
func WithP2PRequest(ctx context.Context) context.Context {
	return context.WithValue(ctx, p2pRequestKey, true)
}

// IsP2PRequest reports whether the context originates from a remote P2P peer request.
func IsP2PRequest(ctx context.Context) bool {
	v, _ := ctx.Value(p2pRequestKey).(bool)
	return v
}

// WithDynamicAllowedTools returns a new context carrying a runtime tool allowlist.
// When non-empty, only listed tools (plus runtime essentials) may execute.
func WithDynamicAllowedTools(ctx context.Context, tools []string) context.Context {
	return context.WithValue(ctx, dynamicAllowedTools, tools)
}

// DynamicAllowedToolsFromContext extracts the runtime tool allowlist from ctx.
// Returns nil if no allowlist is present.
func DynamicAllowedToolsFromContext(ctx context.Context) []string {
	if v, ok := ctx.Value(dynamicAllowedTools).([]string); ok {
		return v
	}
	return nil
}

// WithSpawnDepth returns a new context carrying the current agent spawn depth.
func WithSpawnDepth(ctx context.Context, depth int) context.Context {
	return context.WithValue(ctx, spawnDepthKey, depth)
}

// SpawnDepthFromContext extracts the spawn depth from ctx.
// Returns 0 if no spawn depth is present.
func SpawnDepthFromContext(ctx context.Context) int {
	if v, ok := ctx.Value(spawnDepthKey).(int); ok {
		return v
	}
	return 0
}

// WithSpawnChain returns a new context carrying the agent spawn chain.
// The chain tracks the lineage of spawned agents for cycle detection.
func WithSpawnChain(ctx context.Context, chain []string) context.Context {
	return context.WithValue(ctx, spawnChainKey, chain)
}

// SpawnChainFromContext extracts the spawn chain from ctx.
// Returns nil if no spawn chain is present.
func SpawnChainFromContext(ctx context.Context) []string {
	if v, ok := ctx.Value(spawnChainKey).([]string); ok {
		return v
	}
	return nil
}
