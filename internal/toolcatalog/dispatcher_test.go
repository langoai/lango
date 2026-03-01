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

func TestBuildDispatcher_ReturnsTwo(t *testing.T) {
	tools := BuildDispatcher(setupCatalog())
	require.Len(t, tools, 2)
	assert.Equal(t, "builtin_list", tools[0].Name)
	assert.Equal(t, "builtin_invoke", tools[1].Name)
}

func TestBuiltinList_AllTools(t *testing.T) {
	catalog := setupCatalog()
	tools := BuildDispatcher(catalog)
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
	catalog := setupCatalog()
	tools := BuildDispatcher(catalog)
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
	catalog := setupCatalog()
	tools := BuildDispatcher(catalog)
	invokeTool := tools[1]

	result, err := invokeTool.Handler(context.Background(), map[string]interface{}{
		"tool_name": "exec_shell",
		"params":    map[string]interface{}{"command": "echo hello"},
	})
	require.NoError(t, err)

	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "exec_shell", m["tool"])

	inner, ok := m["result"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "ran: echo hello", inner["stdout"])
}

func TestBuiltinInvoke_NotFound(t *testing.T) {
	catalog := setupCatalog()
	tools := BuildDispatcher(catalog)
	invokeTool := tools[1]

	_, err := invokeTool.Handler(context.Background(), map[string]interface{}{
		"tool_name": "nonexistent_tool",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in catalog")
}

func TestBuiltinInvoke_EmptyToolName(t *testing.T) {
	catalog := setupCatalog()
	tools := BuildDispatcher(catalog)
	invokeTool := tools[1]

	_, err := invokeTool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tool_name is required")
}

func TestBuiltinInvoke_NilParams(t *testing.T) {
	catalog := setupCatalog()
	tools := BuildDispatcher(catalog)
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
	tools := BuildDispatcher(setupCatalog())
	assert.Equal(t, agent.SafetyLevelSafe, tools[0].SafetyLevel, "builtin_list should be safe")
	assert.Equal(t, agent.SafetyLevelDangerous, tools[1].SafetyLevel, "builtin_invoke should be dangerous")
}
