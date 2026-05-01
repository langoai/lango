package toolchain

import (
	"context"

	"github.com/langoai/lango/internal/ctxkeys"
)

// HookContext provides metadata about the current tool execution to hooks.
type HookContext struct {
	ToolName   string
	AgentName  string
	Params     map[string]interface{}
	SessionKey string
	Ctx        context.Context
}

// BlockedToolCall captures structured metadata for a tool invocation blocked by a hook.
type BlockedToolCall struct {
	ToolName    string
	AgentName   string
	SessionKey  string
	BlockReason string
	Params      map[string]interface{}
	Ctx         context.Context
}

// BlockedToolCallSink receives blocked tool call metadata emitted by hook middleware.
type BlockedToolCallSink func(BlockedToolCall)

type blockedToolCallSinkContextKey struct{}

// WithBlockedToolCallSink stores a sink used to observe blocked tool call metadata.
func WithBlockedToolCallSink(ctx context.Context, sink BlockedToolCallSink) context.Context {
	return context.WithValue(ctx, blockedToolCallSinkContextKey{}, sink)
}

func emitBlockedToolCall(ctx context.Context, call BlockedToolCall) {
	sink, ok := ctx.Value(blockedToolCallSinkContextKey{}).(BlockedToolCallSink)
	if !ok || sink == nil {
		return
	}
	sink(call)
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
	// Observe indicates that tool execution should proceed but be logged for review.
	// Commands matching observe-level patterns are legitimate but common obfuscation
	// vectors, so they are allowed with a warning.
	Observe
)

// PreHookResult is returned by pre-hooks to control execution flow.
type PreHookResult struct {
	Action         PreHookAction
	BlockReason    string                 // Used when Action == Block
	ObserveReason  string                 // Used when Action == Observe
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

// WithAgentName delegates to ctxkeys.WithAgentName so that a single canonical
// context key is used across the entire codebase.
var WithAgentName = ctxkeys.WithAgentName

// AgentNameFromContext delegates to ctxkeys.AgentNameFromContext.
var AgentNameFromContext = ctxkeys.AgentNameFromContext
