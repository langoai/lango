package tooloutput

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildTools creates tools for retrieving stored tool output.
func BuildTools(store *OutputStore) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "tool_output_get",
			Description: "Retrieve full or partial stored tool output by reference. Use when a tool result was compressed and you need more detail.",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "system",
				Activity:        agent.ActivityRead,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: agent.Schema().
				Str("ref", "The stored output reference (UUID from _meta.storedRef)").
				Enum("mode", "Retrieval mode: full (default), range, or grep", "full", "range", "grep").
				Int("offset", "Line offset for range mode (0-indexed, default 0)").
				Int("limit", "Max lines to return for range mode (default 100)").
				Str("pattern", "Regex pattern for grep mode").
				Required("ref").
				Build(),
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
