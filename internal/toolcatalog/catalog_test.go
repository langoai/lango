package toolcatalog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

func newTestTool(name string) *agent.Tool {
	return &agent.Tool{
		Name:        name,
		Description: "test tool " + name,
		SafetyLevel: agent.SafetyLevelSafe,
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"tool": name}, nil
		},
	}
}

func TestCatalog_RegisterAndGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		give    []*agent.Tool
		lookup  string
		wantOK  bool
		wantCat string
	}{
		{
			name:    "registered tool found",
			give:    []*agent.Tool{newTestTool("exec_shell")},
			lookup:  "exec_shell",
			wantOK:  true,
			wantCat: "exec",
		},
		{
			name:    "unregistered tool not found",
			give:    []*agent.Tool{newTestTool("exec_shell")},
			lookup:  "nonexistent",
			wantOK:  false,
			wantCat: "",
		},
		{
			name:    "empty catalog",
			give:    nil,
			lookup:  "anything",
			wantOK:  false,
			wantCat: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := New()
			c.RegisterCategory(Category{Name: "exec", Description: "exec tools"})
			c.Register("exec", tt.give)

			entry, ok := c.Get(tt.lookup)
			assert.Equal(t, tt.wantOK, ok)
			if ok {
				assert.Equal(t, tt.wantCat, entry.Category)
				assert.Equal(t, tt.lookup, entry.Tool.Name)
			}
		})
	}
}

func TestCatalog_ListCategories(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "browser", Description: "browser tools", ConfigKey: "tools.browser.enabled", Enabled: true})
	c.RegisterCategory(Category{Name: "exec", Description: "exec tools", ConfigKey: "", Enabled: true})
	c.RegisterCategory(Category{Name: "rag", Description: "RAG tools", ConfigKey: "embedding.rag.enabled", Enabled: false})

	cats := c.ListCategories()
	require.Len(t, cats, 3)

	// Sorted by name.
	assert.Equal(t, "browser", cats[0].Name)
	assert.Equal(t, "exec", cats[1].Name)
	assert.Equal(t, "rag", cats[2].Name)
	assert.False(t, cats[2].Enabled)
}

func TestCatalog_ListTools(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "exec", Description: "exec tools"})
	c.RegisterCategory(Category{Name: "browser", Description: "browser tools"})

	c.Register("exec", []*agent.Tool{newTestTool("exec_shell"), newTestTool("exec_run")})
	c.Register("browser", []*agent.Tool{newTestTool("browser_navigate")})

	tests := []struct {
		name     string
		category string
		wantLen  int
	}{
		{
			name:     "all tools",
			category: "",
			wantLen:  3,
		},
		{
			name:     "exec tools only",
			category: "exec",
			wantLen:  2,
		},
		{
			name:     "browser tools only",
			category: "browser",
			wantLen:  1,
		},
		{
			name:     "nonexistent category",
			category: "nonexistent",
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tools := c.ListTools(tt.category)
			assert.Len(t, tools, tt.wantLen)
		})
	}
}

func TestCatalog_ToolCount(t *testing.T) {
	t.Parallel()

	c := New()
	assert.Equal(t, 0, c.ToolCount())

	c.RegisterCategory(Category{Name: "exec"})
	c.Register("exec", []*agent.Tool{newTestTool("a"), newTestTool("b")})
	assert.Equal(t, 2, c.ToolCount())

	// Re-registering same tool does not increase count.
	c.Register("exec", []*agent.Tool{newTestTool("a")})
	assert.Equal(t, 2, c.ToolCount())
}

func TestCatalog_GetToolSafetyLevel(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "exec", Description: "exec tools"})

	safeTool := &agent.Tool{
		Name:        "read_file",
		Description: "read a file",
		SafetyLevel: agent.SafetyLevelSafe,
	}
	dangerousTool := &agent.Tool{
		Name:        "exec_shell",
		Description: "execute shell command",
		SafetyLevel: agent.SafetyLevelDangerous,
	}
	c.Register("exec", []*agent.Tool{safeTool, dangerousTool})

	tests := []struct {
		name      string
		give      string
		wantLevel agent.SafetyLevel
		wantOK    bool
	}{
		{
			name:      "safe tool found",
			give:      "read_file",
			wantLevel: agent.SafetyLevelSafe,
			wantOK:    true,
		},
		{
			name:      "dangerous tool found",
			give:      "exec_shell",
			wantLevel: agent.SafetyLevelDangerous,
			wantOK:    true,
		},
		{
			name:      "unknown tool returns dangerous",
			give:      "nonexistent",
			wantLevel: agent.SafetyLevelDangerous,
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			level, ok := c.GetToolSafetyLevel(tt.give)
			assert.Equal(t, tt.wantLevel, level)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestCatalog_InsertionOrder(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "a"})
	c.RegisterCategory(Category{Name: "b"})

	c.Register("a", []*agent.Tool{newTestTool("z_tool")})
	c.Register("b", []*agent.Tool{newTestTool("a_tool")})

	tools := c.ListTools("")
	require.Len(t, tools, 2)
	assert.Equal(t, "z_tool", tools[0].Name, "insertion order preserved")
	assert.Equal(t, "a_tool", tools[1].Name, "insertion order preserved")
}
