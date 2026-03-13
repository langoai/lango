package toolchain

import (
	"context"
	"encoding/json"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/logging"
)

const defaultMaxOutputChars = 8000

// WithTruncate returns a middleware that caps tool result text size.
// Results exceeding maxChars are truncated with a marker.
func WithTruncate(maxChars int) Middleware {
	if maxChars <= 0 {
		maxChars = defaultMaxOutputChars
	}
	return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			result, err := next(ctx, params)
			if err != nil {
				return result, err
			}
			return truncateResult(result, maxChars, tool.Name), nil
		}
	}
}

// truncateResult limits the text size of a tool result.
func truncateResult(result interface{}, maxChars int, toolName string) interface{} {
	var text string
	switch v := result.(type) {
	case string:
		text = v
	case nil:
		return result
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return result
		}
		text = string(data)
	}

	if len(text) <= maxChars {
		return result
	}

	logging.App().Warnw("tool output truncated",
		"tool", toolName,
		"original", len(text),
		"limit", maxChars)

	truncated := text[:maxChars] + "\n... [output truncated]"

	// If the original was a string, return truncated string.
	// Otherwise return the truncated JSON string representation.
	if _, ok := result.(string); ok {
		return truncated
	}
	return truncated
}
