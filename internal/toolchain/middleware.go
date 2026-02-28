package toolchain

import (
	"github.com/langoai/lango/internal/agent"
)

// Middleware wraps a tool handler. It receives the tool (for metadata access) and the next handler.
type Middleware func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler

// Chain applies middlewares to a single tool, returning a new tool with wrapped handler.
// Middlewares are applied in order: first middleware is outermost (executed first).
func Chain(tool *agent.Tool, middlewares ...Middleware) *agent.Tool {
	if len(middlewares) == 0 {
		return tool
	}
	// Build from inside out: last middleware wraps original, first middleware is outermost.
	handler := tool.Handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](tool, handler)
	}
	return &agent.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		Parameters:  tool.Parameters,
		SafetyLevel: tool.SafetyLevel,
		Handler:     handler,
	}
}

// ChainAll applies the same middleware stack to all tools.
func ChainAll(tools []*agent.Tool, middlewares ...Middleware) []*agent.Tool {
	if len(middlewares) == 0 {
		return tools
	}
	result := make([]*agent.Tool, len(tools))
	for i, t := range tools {
		result[i] = Chain(t, middlewares...)
	}
	return result
}
