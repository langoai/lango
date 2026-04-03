package toolcatalog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

func setupCatalog() *Catalog {
	c := New()
	c.RegisterCategory(Category{Name: "exec", Description: "exec tools", Enabled: true})
	c.RegisterCategory(Category{Name: "browser", Description: "browser tools", ConfigKey: "tools.browser.enabled", Enabled: true})

	c.Register("exec", []*agent.Tool{
		{
			Name:        "exec_shell",
			Description: "execute a shell command",
			SafetyLevel: agent.SafetyLevelDangerous,
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				cmd, _ := params["command"].(string)
				return map[string]interface{}{"stdout": "ran: " + cmd}, nil
			},
		},
	})
	c.Register("browser", []*agent.Tool{
		{
			Name:        "browser_navigate",
			Description: "navigate to a URL",
			SafetyLevel: agent.SafetyLevelSafe,
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				url, _ := params["url"].(string)
				return map[string]interface{}{"navigated": url}, nil
			},
		},
	})
	return c
}

// setupCatalogAndIndex builds both a test catalog and its search index.
func setupCatalogAndIndex() (*Catalog, *SearchIndex) {
	c := setupCatalog()
	return c, NewSearchIndex(c)
}

func TestBuildDispatcher_ReturnsFour(t *testing.T) {
	t.Parallel()

	catalog, index := setupCatalogAndIndex()
	tools := BuildDispatcher(catalog, index)
	require.Len(t, tools, 4)
	assert.Equal(t, "builtin_list", tools[0].Name)
	assert.Equal(t, "builtin_invoke", tools[1].Name)
	assert.Equal(t, "builtin_health", tools[2].Name)
	assert.Equal(t, "builtin_search", tools[3].Name)
}

func TestBuiltinList_AllTools(t *testing.T) {
	t.Parallel()

	catalog, index := setupCatalogAndIndex()
	tools := BuildDispatcher(catalog, index)
	listTool := tools[0]

	result, err := listTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)

	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 2, m["total"])

	toolList, ok := m["tools"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, toolList, 2)
}

func TestBuiltinList_FilterByCategory(t *testing.T) {
	t.Parallel()

	catalog, index := setupCatalogAndIndex()
	tools := BuildDispatcher(catalog, index)
	listTool := tools[0]

	result, err := listTool.Handler(context.Background(), map[string]interface{}{
		"category": "exec",
	})
	require.NoError(t, err)

	m, ok := result.(map[string]interface{})
	require.True(t, ok)

	toolList, ok := m["tools"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, toolList, 1)
	assert.Equal(t, "exec_shell", toolList[0]["name"])
}

func TestBuiltinInvoke_Success(t *testing.T) {
	t.Parallel()

	catalog, index := setupCatalogAndIndex()
	tools := BuildDispatcher(catalog, index)
	invokeTool := tools[1]

	// Use a safe tool (browser_navigate) — dangerous tools are blocked.
	result, err := invokeTool.Handler(context.Background(), map[string]interface{}{
		"tool_name": "browser_navigate",
		"params":    map[string]interface{}{"url": "https://example.com"},
	})
	require.NoError(t, err)

	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "browser_navigate", m["tool"])

	inner, ok := m["result"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "https://example.com", inner["navigated"])
}

func TestBuiltinInvoke_BlocksDangerousTools(t *testing.T) {
	t.Parallel()

	catalog, index := setupCatalogAndIndex()
	tools := BuildDispatcher(catalog, index)
	invokeTool := tools[1]

	_, err := invokeTool.Handler(context.Background(), map[string]interface{}{
		"tool_name": "exec_shell",
		"params":    map[string]interface{}{"command": "echo hello"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires approval")
	assert.Contains(t, err.Error(), "delegate to the appropriate sub-agent")
}

func TestBuiltinInvoke_NotFound(t *testing.T) {
	t.Parallel()

	catalog, index := setupCatalogAndIndex()
	tools := BuildDispatcher(catalog, index)
	invokeTool := tools[1]

	_, err := invokeTool.Handler(context.Background(), map[string]interface{}{
		"tool_name": "nonexistent_tool",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in catalog")
}

func TestBuiltinInvoke_EmptyToolName(t *testing.T) {
	t.Parallel()

	catalog, index := setupCatalogAndIndex()
	tools := BuildDispatcher(catalog, index)
	invokeTool := tools[1]

	_, err := invokeTool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tool_name is required")
}

func TestBuiltinInvoke_NilParams(t *testing.T) {
	t.Parallel()

	catalog, index := setupCatalogAndIndex()
	tools := BuildDispatcher(catalog, index)
	invokeTool := tools[1]

	// Invoke without params — handler should receive empty map.
	result, err := invokeTool.Handler(context.Background(), map[string]interface{}{
		"tool_name": "browser_navigate",
	})
	require.NoError(t, err)

	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "browser_navigate", m["tool"])
}

func TestDispatcher_SafetyLevels(t *testing.T) {
	t.Parallel()

	catalog, index := setupCatalogAndIndex()
	tools := BuildDispatcher(catalog, index)
	assert.Equal(t, agent.SafetyLevelSafe, tools[0].SafetyLevel, "builtin_list should be safe")
	assert.Equal(t, agent.SafetyLevelDangerous, tools[1].SafetyLevel, "builtin_invoke should be dangerous")
	assert.Equal(t, agent.SafetyLevelSafe, tools[2].SafetyLevel, "builtin_health should be safe")
	assert.Equal(t, agent.SafetyLevelSafe, tools[3].SafetyLevel, "builtin_search should be safe")
}

func TestBuiltinHealth_ShowsDisabledCategories(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "exec", Description: "exec tools", Enabled: true})
	c.RegisterCategory(Category{Name: "smartaccount", Description: "ERC-7579 (disabled)", ConfigKey: "smartAccount.enabled", Enabled: false})
	c.Register("exec", []*agent.Tool{
		{
			Name:        "exec_shell",
			Description: "execute a shell command",
			SafetyLevel: agent.SafetyLevelDangerous,
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				return nil, nil
			},
		},
	})

	idx := NewSearchIndex(c)
	tools := BuildDispatcher(c, idx)
	healthTool := tools[2]

	result, err := healthTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)

	m, ok := result.(map[string]interface{})
	require.True(t, ok)

	enabled, ok := m["enabled_categories"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, enabled, 1)
	assert.Equal(t, "exec", enabled[0]["name"])

	// Verify enabled categories include tool names.
	toolNames, ok := enabled[0]["tools"].([]string)
	require.True(t, ok)
	assert.Equal(t, []string{"exec_shell"}, toolNames)

	disabled, ok := m["disabled_categories"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, disabled, 1)
	assert.Equal(t, "smartaccount", disabled[0]["name"])
	assert.Contains(t, disabled[0]["hint"], "lango config set smartAccount.enabled true")
}

func TestBuiltinHealth_ToolNamesPerCategory(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "cron", Description: "cron tools", ConfigKey: "cron.enabled", Enabled: true})
	c.Register("cron", []*agent.Tool{
		{Name: "cron_add", Description: "add cron job", SafetyLevel: agent.SafetyLevelSafe,
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) { return nil, nil }},
		{Name: "cron_list", Description: "list cron jobs", SafetyLevel: agent.SafetyLevelSafe,
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) { return nil, nil }},
		{Name: "cron_remove", Description: "remove cron job", SafetyLevel: agent.SafetyLevelDangerous,
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) { return nil, nil }},
	})

	idx := NewSearchIndex(c)
	tools := BuildDispatcher(c, idx)
	healthTool := tools[2]

	result, err := healthTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	enabled := m["enabled_categories"].([]map[string]interface{})
	require.Len(t, enabled, 1)

	toolNames := enabled[0]["tools"].([]string)
	assert.Equal(t, []string{"cron_add", "cron_list", "cron_remove"}, toolNames)
	assert.Equal(t, 3, enabled[0]["tool_count"])
}
