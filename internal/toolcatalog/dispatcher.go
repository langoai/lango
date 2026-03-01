package toolcatalog

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
)

// BuildDispatcher returns two meta-tools that provide dynamic access to
// the catalog: builtin_list (discovery) and builtin_invoke (proxy execution).
func BuildDispatcher(catalog *Catalog) []*agent.Tool {
	return []*agent.Tool{
		buildListTool(catalog),
		buildInvokeTool(catalog),
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

			tools := catalog.ListTools(category)
			toolSummaries := make([]map[string]interface{}, 0, len(tools))
			for _, t := range tools {
				toolSummaries = append(toolSummaries, map[string]interface{}{
					"name":         t.Name,
					"description":  t.Description,
					"category":     t.Category,
					"safety_level": t.SafetyLevel,
				})
			}

			return map[string]interface{}{
				"categories": catSummaries,
				"tools":      toolSummaries,
				"total":      catalog.ToolCount(),
			}, nil
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
