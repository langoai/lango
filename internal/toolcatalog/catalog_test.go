package toolcatalog

import (
	"context"
	"strconv"
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

func newTestToolWithCapability(name string, cap agent.ToolCapability) *agent.Tool {
	t := newTestTool(name)
	t.Capability = cap
	return t
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

func TestCatalog_ListTools_CapabilityFields(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "fs"})

	c.Register("fs", []*agent.Tool{
		newTestToolWithCapability("fs_read", agent.ToolCapability{
			Aliases:              []string{"cat", "read"},
			SearchHints:          []string{"file", "content"},
			Exposure:             agent.ExposureDeferred,
			ReadOnly:             true,
			Activity:             agent.ActivityRead,
			RequiredCapabilities: []string{"filesystem"},
		}),
		newTestToolWithCapability("fs_write", agent.ToolCapability{
			Activity: agent.ActivityWrite,
		}),
	})

	schemas := c.ListTools("")
	require.Len(t, schemas, 2)

	// fs_read: all capability fields populated.
	s := schemas[0]
	assert.Equal(t, "fs_read", s.Name)
	assert.Equal(t, []string{"cat", "read"}, s.Aliases)
	assert.Equal(t, []string{"file", "content"}, s.SearchHints)
	assert.Equal(t, "deferred", s.Exposure)
	assert.True(t, s.ReadOnly)
	assert.Equal(t, "read", s.Activity)
	assert.Equal(t, []string{"filesystem"}, s.RequiredCapabilities)

	// fs_write: most capability fields at zero value.
	s2 := schemas[1]
	assert.Equal(t, "fs_write", s2.Name)
	assert.Nil(t, s2.Aliases)
	assert.Nil(t, s2.SearchHints)
	assert.Equal(t, "", s2.Exposure, "ExposureDefault omits exposure")
	assert.False(t, s2.ReadOnly)
	assert.Equal(t, "write", s2.Activity)
	assert.Nil(t, s2.RequiredCapabilities)
}

func TestCatalog_ListVisibleTools(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "core"})
	c.RegisterCategory(Category{Name: "ext"})

	c.Register("core", []*agent.Tool{
		newTestToolWithCapability("visible_default", agent.ToolCapability{
			Exposure: agent.ExposureDefault,
		}),
		newTestToolWithCapability("visible_always", agent.ToolCapability{
			Exposure: agent.ExposureAlwaysVisible,
		}),
		newTestToolWithCapability("deferred_tool", agent.ToolCapability{
			Exposure: agent.ExposureDeferred,
		}),
		newTestToolWithCapability("hidden_tool", agent.ToolCapability{
			Exposure: agent.ExposureHidden,
		}),
	})
	c.Register("ext", []*agent.Tool{
		newTestToolWithCapability("ext_visible", agent.ToolCapability{
			Exposure: agent.ExposureDefault,
		}),
		newTestToolWithCapability("ext_deferred", agent.ToolCapability{
			Exposure: agent.ExposureDeferred,
		}),
	})

	tests := []struct {
		name      string
		category  string
		wantNames []string
	}{
		{
			name:      "all visible tools across categories",
			category:  "",
			wantNames: []string{"visible_default", "visible_always", "ext_visible"},
		},
		{
			name:      "visible tools in core category",
			category:  "core",
			wantNames: []string{"visible_default", "visible_always"},
		},
		{
			name:      "visible tools in ext category",
			category:  "ext",
			wantNames: []string{"ext_visible"},
		},
		{
			name:      "nonexistent category returns empty",
			category:  "nonexistent",
			wantNames: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schemas := c.ListVisibleTools(tt.category)
			var names []string
			for _, s := range schemas {
				names = append(names, s.Name)
			}
			assert.Equal(t, tt.wantNames, names)
		})
	}
}

func TestCatalog_SearchableEntries(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "tools"})

	c.Register("tools", []*agent.Tool{
		newTestToolWithCapability("default_tool", agent.ToolCapability{
			Exposure: agent.ExposureDefault,
		}),
		newTestToolWithCapability("always_tool", agent.ToolCapability{
			Exposure: agent.ExposureAlwaysVisible,
		}),
		newTestToolWithCapability("deferred_tool", agent.ToolCapability{
			Exposure: agent.ExposureDeferred,
		}),
		newTestToolWithCapability("hidden_tool", agent.ToolCapability{
			Exposure: agent.ExposureHidden,
		}),
	})

	entries := c.SearchableEntries()

	var names []string
	for _, e := range entries {
		names = append(names, e.Tool.Name)
	}

	// Hidden excluded, all others included.
	assert.Equal(t, []string{"default_tool", "always_tool", "deferred_tool"}, names)
}

func TestCatalog_SearchableEntries_Empty(t *testing.T) {
	t.Parallel()

	c := New()
	entries := c.SearchableEntries()
	assert.Empty(t, entries)
}

func TestCatalog_DeferredToolCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		exposures []agent.ExposurePolicy
		wantCount int
	}{
		{
			name:      "no tools",
			exposures: nil,
			wantCount: 0,
		},
		{
			name:      "no deferred tools",
			exposures: []agent.ExposurePolicy{agent.ExposureDefault, agent.ExposureHidden},
			wantCount: 0,
		},
		{
			name: "mixed exposures",
			exposures: []agent.ExposurePolicy{
				agent.ExposureDefault,
				agent.ExposureDeferred,
				agent.ExposureHidden,
				agent.ExposureDeferred,
				agent.ExposureAlwaysVisible,
			},
			wantCount: 2,
		},
		{
			name:      "all deferred",
			exposures: []agent.ExposurePolicy{agent.ExposureDeferred, agent.ExposureDeferred},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := New()
			c.RegisterCategory(Category{Name: "cat"})
			var tools []*agent.Tool
			for i, exp := range tt.exposures {
				tools = append(tools, newTestToolWithCapability(
					"tool_"+strconv.Itoa(i),
					agent.ToolCapability{Exposure: exp},
				))
			}
			c.Register("cat", tools)
			assert.Equal(t, tt.wantCount, c.DeferredToolCount())
		})
	}
}

func TestSaveableToolNames(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "test", Enabled: true})
	c.Register("test", []*agent.Tool{
		newTestToolWithCapability("reader", agent.ToolCapability{ReadOnly: true}),
		newTestToolWithCapability("querier", agent.ToolCapability{Activity: agent.ActivityQuery}),
		newTestToolWithCapability("writer", agent.ToolCapability{Activity: agent.ActivityWrite}),
		newTestToolWithCapability("executor", agent.ToolCapability{Activity: agent.ActivityExecute}),
		newTestToolWithCapability("plain", agent.ToolCapability{}),
	})

	names := c.SaveableToolNames()
	assert.Equal(t, []string{"querier", "reader"}, names)
}
