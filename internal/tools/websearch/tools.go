package websearch

import (
	"context"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildTools returns web search tools that use HTTP-only DuckDuckGo search
// without requiring a browser session.
func BuildTools() []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "web_search",
			Description: "Search the web for information. Returns structured search results without requiring a browser session.",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:    "web",
				Aliases:     []string{"search_web", "internet_search"},
				SearchHints: []string{"search", "find", "lookup", "web"},
				ReadOnly:    true,
				ConcurrencySafe: true,
				Activity:    agent.ActivityQuery,
			},
			Parameters: agent.Schema().
				Str("query", "The search query to run").
				Int("limit", "Maximum number of results to return (default: 5, max: 20)").
				Required("query").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				query, err := toolparam.RequireString(params, "query")
				if err != nil {
					return nil, err
				}

				limit := toolparam.OptionalInt(params, "limit", defaultLimit)

				results, err := Search(ctx, query, limit)
				if err != nil {
					return nil, err
				}

				return map[string]interface{}{
					"query":   query,
					"results": results,
					"count":   len(results),
				}, nil
			},
		},
	}
}
