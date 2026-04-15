package toolcatalog

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/agent"
)

func newModeTestCatalog(t *testing.T) *Catalog {
	t.Helper()
	c := New()
	c.RegisterCategory(Category{Name: "exec", Description: "exec tools", Enabled: true})
	c.RegisterCategory(Category{Name: "web", Description: "web tools", Enabled: true})
	c.RegisterCategory(Category{Name: "file", Description: "file tools", Enabled: true})

	c.Register("exec", []*agent.Tool{
		newTestTool("shell_exec"),
		newTestTool("run_process"),
	})
	c.Register("web", []*agent.Tool{
		newTestTool("web_search"),
		newTestTool("web_fetch"),
	})
	c.Register("file", []*agent.Tool{
		newTestTool("read_file"),
		newTestTool("write_file"),
	})
	return c
}

func TestResolveModeAllowlist_ExpandsCategoryRefs(t *testing.T) {
	c := newModeTestCatalog(t)
	allow := c.ResolveModeAllowlist([]string{"@exec", "web_search"})
	assert.True(t, allow["shell_exec"], "@exec should include shell_exec")
	assert.True(t, allow["run_process"], "@exec should include run_process")
	assert.True(t, allow["web_search"], "explicit tool name included")
	assert.False(t, allow["web_fetch"], "only web_search, not full @web")
	assert.False(t, allow["read_file"], "not in allowlist")
}

func TestResolveModeAllowlist_EmptyReturnsEmpty(t *testing.T) {
	c := newModeTestCatalog(t)
	allow := c.ResolveModeAllowlist(nil)
	assert.Empty(t, allow)
}

func TestResolveModeAllowlist_UnknownCategoryEmpty(t *testing.T) {
	c := newModeTestCatalog(t)
	allow := c.ResolveModeAllowlist([]string{"@nonexistent"})
	assert.Empty(t, allow)
}

func TestListVisibleToolsForMode_FiltersToAllowlist(t *testing.T) {
	c := newModeTestCatalog(t)
	schemas := c.ListVisibleToolsForMode([]string{"@exec"})
	names := make(map[string]bool)
	for _, s := range schemas {
		names[s.Name] = true
	}
	assert.True(t, names["shell_exec"])
	assert.True(t, names["run_process"])
	assert.False(t, names["web_search"], "web tools filtered out")
}

func TestListVisibleToolsForMode_EmptyModeEquivalentToListVisible(t *testing.T) {
	c := newModeTestCatalog(t)
	withoutMode := c.ListVisibleTools("")
	emptyMode := c.ListVisibleToolsForMode(nil)
	assert.Equal(t, len(withoutMode), len(emptyMode))
}

func TestListVisibleToolsForMode_MixedExplicitAndCategory(t *testing.T) {
	c := newModeTestCatalog(t)
	schemas := c.ListVisibleToolsForMode([]string{"@web", "read_file"})
	names := make(map[string]bool)
	for _, s := range schemas {
		names[s.Name] = true
	}
	assert.True(t, names["web_search"])
	assert.True(t, names["web_fetch"])
	assert.True(t, names["read_file"])
	assert.False(t, names["write_file"])
	assert.False(t, names["shell_exec"])
}
