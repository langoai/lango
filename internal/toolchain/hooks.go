package toolchain

import "context"

// HookContext provides metadata about the current tool execution to hooks.
type HookContext struct {
	ToolName   string
	AgentName  string
	Params     map[string]interface{}
	SessionKey string
	Ctx        context.Context
}

// PreHookAction determines what happens after a pre-hook runs.
type PreHookAction int

const (
	// Continue indicates that tool execution should proceed normally.
	Continue PreHookAction = iota
	// Block indicates that tool execution should be stopped.
	Block
	// Modify indicates that tool execution should proceed with modified params.
	Modify
)

// PreHookResult is returned by pre-hooks to control execution flow.
type PreHookResult struct {
	Action         PreHookAction
	BlockReason    string                 // Used when Action == Block
	ModifiedParams map[string]interface{} // Used when Action == Modify
}

// PreToolHook runs before tool execution.
type PreToolHook interface {
	Name() string
	Priority() int // Lower = runs first
	Pre(ctx HookContext) (PreHookResult, error)
}

// PostToolHook runs after tool execution.
type PostToolHook interface {
	Name() string
	Priority() int // Lower = runs first
	Post(ctx HookContext, result interface{}, toolErr error) error
}

// contextKey is a private type to avoid collisions with other packages.
type contextKey string

const agentNameCtxKey contextKey = "toolchain.agent_name"

// WithAgentName sets the agent name in context (called by ADK adapter).
func WithAgentName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, agentNameCtxKey, name)
}

// AgentNameFromContext extracts the agent name from context.
func AgentNameFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(agentNameCtxKey).(string); ok {
		return v
	}
	return ""
}
