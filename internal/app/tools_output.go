package app

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/tooloutput"
	"github.com/langoai/lango/internal/toolparam"
)

// buildOutputTools creates tools for retrieving stored tool output.
func buildOutputTools(store *tooloutput.OutputStore) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "tool_output_get",
			Description: "Retrieve full or partial stored tool output by reference. Use when a tool result was compressed and you need more detail.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ref":     map[string]interface{}{"type": "string", "description": "The stored output reference (UUID from _meta.storedRef)"},
					"mode":    map[string]interface{}{"type": "string", "description": "Retrieval mode: full (default), range, or grep", "enum": []string{"full", "range", "grep"}},
					"offset":  map[string]interface{}{"type": "integer", "description": "Line offset for range mode (0-indexed, default 0)"},
					"limit":   map[string]interface{}{"type": "integer", "description": "Max lines to return for range mode (default 100)"},
					"pattern": map[string]interface{}{"type": "string", "description": "Regex pattern for grep mode"},
				},
				"required": []string{"ref"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				ref, err := toolparam.RequireString(params, "ref")
				if err != nil {
					return nil, err
				}

				mode := toolparam.OptionalString(params, "mode", "full")

				switch mode {
				case "range":
					offset := toolparam.OptionalInt(params, "offset", 0)
					limit := toolparam.OptionalInt(params, "limit", 100)
					content, total, ok := store.GetRange(ref, offset, limit)
					if !ok {
						return nil, fmt.Errorf("output ref %q not found or expired", ref)
					}
					return toolparam.Response{
						"content":    content,
						"totalLines": total,
						"offset":     offset,
						"limit":      limit,
					}, nil

				case "grep":
					pattern, pErr := toolparam.RequireString(params, "pattern")
					if pErr != nil {
						return nil, fmt.Errorf("pattern required for grep mode")
					}
					content, ok := store.Grep(ref, pattern)
					if !ok {
						return nil, fmt.Errorf("output ref %q not found or expired", ref)
					}
					return toolparam.Response{
						"matches": content,
					}, nil

				default: // "full"
					content, ok := store.Get(ref)
					if !ok {
						return nil, fmt.Errorf("output ref %q not found or expired", ref)
					}
					return toolparam.Response{
						"content": content,
					}, nil
				}
			},
		},
	}
}
