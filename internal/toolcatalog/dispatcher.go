package toolcatalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/agent"
)

// BuildDispatcher returns meta-tools that provide dynamic access to
// the catalog: builtin_list (discovery), builtin_invoke (proxy execution),
// builtin_health (diagnostics), and builtin_search (keyword search).
func BuildDispatcher(catalog *Catalog, index *SearchIndex) []*agent.Tool {
	return []*agent.Tool{
		buildListTool(catalog),
		buildInvokeTool(catalog),
		buildHealthTool(catalog),
		buildSearchTool(index),
	}
}

// buildListTool creates the builtin_list tool for discovering registered tools.
func buildListTool(catalog *Catalog) *agent.Tool {
	return &agent.Tool{
		Name: "builtin_list",
		Description: "List available built-in tools and categories. " +
			"Use this to discover what tools are registered in the system.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Optional category filter. If omitted, all tools are listed.",
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			category, _ := params["category"].(string)

			categories := catalog.ListCategories()
			catSummaries := make([]map[string]interface{}, 0, len(categories))
			for _, cat := range categories {
				catSummaries = append(catSummaries, map[string]interface{}{
					"name":        cat.Name,
					"description": cat.Description,
					"config_key":  cat.ConfigKey,
					"enabled":     cat.Enabled,
				})
			}

			tools := catalog.ListVisibleTools(category)
			toolSummaries := make([]map[string]interface{}, 0, len(tools))
			for _, t := range tools {
				toolSummaries = append(toolSummaries, map[string]interface{}{
					"name":         t.Name,
					"description":  t.Description,
					"category":     t.Category,
					"safety_level": t.SafetyLevel,
				})
			}

			result := map[string]interface{}{
				"categories":     catSummaries,
				"tools":          toolSummaries,
				"total":          catalog.ToolCount(),
				"deferred_count": catalog.DeferredToolCount(),
			}

			if catalog.DeferredToolCount() > 0 {
				result["hint"] = "Use builtin_search to discover additional tools not shown here."
			}

			return result, nil
		},
	}
}

// buildInvokeTool creates the builtin_invoke tool for proxy-executing catalog tools.
func buildInvokeTool(catalog *Catalog) *agent.Tool {
	return &agent.Tool{
		Name: "builtin_invoke",
		Description: "Invoke a registered built-in tool by name. " +
			"Use builtin_list to discover available tools first.",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tool_name": map[string]interface{}{
					"type":        "string",
					"description": "The name of the built-in tool to invoke",
				},
				"params": map[string]interface{}{
					"type":        "object",
					"description": "Parameters to pass to the tool",
				},
			},
			"required": []string{"tool_name"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			toolName, _ := params["tool_name"].(string)
			if toolName == "" {
				return nil, fmt.Errorf("tool_name is required")
			}

			entry, ok := catalog.Get(toolName)
			if !ok {
				return nil, fmt.Errorf("tool %q not found in catalog", toolName)
			}

			// Block dangerous tools from being invoked via the dispatcher.
			// Dangerous tools must be executed through their owning sub-agent
			// which has the proper approval chain wired by ADK middleware.
			if entry.Tool.SafetyLevel >= agent.SafetyLevelDangerous {
				return nil, fmt.Errorf(
					"tool %q requires approval (safety=%s); delegate to the appropriate sub-agent instead",
					toolName, entry.Tool.SafetyLevel,
				)
			}

			toolParams, _ := params["params"].(map[string]interface{})
			if toolParams == nil {
				toolParams = make(map[string]interface{})
			}

			result, err := entry.Tool.Handler(ctx, toolParams)
			if err != nil {
				return nil, fmt.Errorf("invoke %q: %w", toolName, err)
			}

			return map[string]interface{}{
				"tool":   toolName,
				"result": result,
			}, nil
		},
	}
}

// buildHealthTool creates the builtin_health tool for diagnosing tool registration status.
func buildHealthTool(catalog *Catalog) *agent.Tool {
	return &agent.Tool{
		Name: "builtin_health",
		Description: "Diagnose tool registration status. Shows all categories (enabled and disabled) " +
			"with tool names and required config keys. Use this when tools appear to be missing.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			categories := catalog.ListCategories()

			var enabled []map[string]interface{}
			var disabled []map[string]interface{}

			for _, cat := range categories {
				entry := map[string]interface{}{
					"name":        cat.Name,
					"description": cat.Description,
				}
				if cat.ConfigKey != "" {
					entry["config_key"] = cat.ConfigKey
				}

				if cat.Enabled {
					toolNames := catalog.ToolNamesForCategory(cat.Name)
					entry["tool_count"] = len(toolNames)
					entry["tools"] = toolNames
					enabled = append(enabled, entry)
				} else {
					entry["hint"] = fmt.Sprintf(
						"Enable with: lango config set %s true", cat.ConfigKey,
					)
					disabled = append(disabled, entry)
				}
			}

			return map[string]interface{}{
				"enabled_categories":  enabled,
				"disabled_categories": disabled,
				"total_tools":         catalog.ToolCount(),
			}, nil
		},
	}
}

// buildSearchTool creates the builtin_search tool for keyword-based tool discovery.
func buildSearchTool(index *SearchIndex) *agent.Tool {
	return &agent.Tool{
		Name: "builtin_search",
		Description: "Search for tools by keyword, capability, or description. " +
			"Returns ranked results with relevance scores.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query: keywords, tool names, or capability descriptions.",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results to return. Defaults to 10.",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Optional category filter to narrow search results.",
				},
				"activity": map[string]interface{}{
					"type":        "string",
					"description": "Optional activity kind filter (e.g. read, write, execute).",
				},
			},
			"required": []string{"query"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			query, _ := params["query"].(string)
			if query == "" {
				return nil, fmt.Errorf("query is required")
			}

			limit := 10
			if v, ok := params["limit"].(float64); ok && v > 0 {
				limit = int(v)
			}

			// Search without limit first, then filter, then apply limit.
			results := index.Search(query, 0)

			// Apply optional category filter.
			if cat, _ := params["category"].(string); cat != "" {
				catLower := strings.ToLower(cat)
				filtered := results[:0]
				for _, r := range results {
					if strings.ToLower(r.Category) == catLower {
						filtered = append(filtered, r)
					}
				}
				results = filtered
			}

			// Apply optional activity filter.
			if act, _ := params["activity"].(string); act != "" {
				actLower := strings.ToLower(act)
				filtered := results[:0]
				for _, r := range results {
					if strings.ToLower(r.Activity) == actLower {
						filtered = append(filtered, r)
					}
				}
				results = filtered
			}

			// Apply limit after filtering.
			if len(results) > limit {
				results = results[:limit]
			}

			resultMaps := make([]map[string]interface{}, 0, len(results))
			for _, r := range results {
				resultMaps = append(resultMaps, map[string]interface{}{
					"name":        r.Name,
					"description": r.Description,
					"category":    r.Category,
					"score":       r.Score,
					"match_field": r.MatchField,
				})
			}

			return map[string]interface{}{
				"query":   query,
				"results": resultMaps,
				"count":   len(resultMaps),
			}, nil
		},
	}
}
