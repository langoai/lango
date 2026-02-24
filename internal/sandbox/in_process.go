package sandbox

import "context"

// ToolFunc is the signature for a tool handler function.
type ToolFunc func(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error)

// InProcessExecutor delegates directly to a ToolFunc without isolation.
// Use this for local/trusted tool executions where isolation is unnecessary.
type InProcessExecutor struct {
	fn ToolFunc
}

// NewInProcessExecutor wraps an existing tool function as an Executor.
func NewInProcessExecutor(fn ToolFunc) *InProcessExecutor {
	return &InProcessExecutor{fn: fn}
}

// Execute runs the tool in the current process.
func (e *InProcessExecutor) Execute(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
	return e.fn(ctx, toolName, params)
}
