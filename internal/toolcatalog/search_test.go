package toolcatalog

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

// buildTestCatalog creates a catalog with diverse tools for search testing.
func buildTestCatalog() *Catalog {
	c := New()
	c.RegisterCategory(Category{Name: "filesystem", Description: "filesystem tools", Enabled: true})
	c.RegisterCategory(Category{Name: "exec", Description: "execution tools", Enabled: true})
	c.RegisterCategory(Category{Name: "browser", Description: "browser tools", Enabled: true})

	c.Register("filesystem", []*agent.Tool{
		newTestToolWithCapability("fs_read", agent.ToolCapability{
			Aliases:     []string{"cat", "read"},
			Category:    "filesystem",
			SearchHints: []string{"file", "content", "read"},
			Exposure:    agent.ExposureDefault,
			ReadOnly:    true,
			Activity:    agent.ActivityRead,
		}),
		newTestToolWithCapability("fs_write", agent.ToolCapability{
			Aliases:     []string{"write", "save"},
			Category:    "filesystem",
			SearchHints: []string{"file", "write", "create"},
			Exposure:    agent.ExposureDefault,
			Activity:    agent.ActivityWrite,
		}),
		newTestToolWithCapability("fs_list", agent.ToolCapability{
			Aliases:     []string{"ls", "dir"},
			Category:    "filesystem",
			SearchHints: []string{"directory", "listing"},
			Exposure:    agent.ExposureDeferred,
			ReadOnly:    true,
			Activity:    agent.ActivityRead,
		}),
	})

	c.Register("exec", []*agent.Tool{
		newTestToolWithCapability("exec_shell", agent.ToolCapability{
			Aliases:     []string{"shell", "bash", "sh"},
			Category:    "execution",
			SearchHints: []string{"command", "terminal", "run"},
			Exposure:    agent.ExposureDefault,
			Activity:    agent.ActivityExecute,
		}),
	})

	c.Register("browser", []*agent.Tool{
		newTestToolWithCapability("browser_navigate", agent.ToolCapability{
			Aliases:     []string{"goto", "open"},
			Category:    "web",
			SearchHints: []string{"url", "page", "navigate"},
			Exposure:    agent.ExposureDefault,
			Activity:    agent.ActivityRead,
		}),
		// Hidden tool — must NOT appear in search results.
		newTestToolWithCapability("browser_internal", agent.ToolCapability{
			Exposure: agent.ExposureHidden,
		}),
	})

	return c
}

func TestSearch_ExactNameOutranksDescription(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	results := idx.Search("fs_read", 0)
	require.NotEmpty(t, results)
	assert.Equal(t, "fs_read", results[0].Name)
	assert.Equal(t, weightExactName, results[0].Score)
	assert.Equal(t, "name", results[0].MatchField)

	// Any tool that only matches "fs_read" via description should rank lower.
	for i := 1; i < len(results); i++ {
		assert.Less(t, results[i].Score, results[0].Score)
	}
}

func TestSearch_NamePrefixMatch(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	results := idx.Search("fs_", 0)
	require.GreaterOrEqual(t, len(results), 3)

	// All fs_ tools should appear with prefix score.
	for _, r := range results[:3] {
		assert.Contains(t, r.Name, "fs_")
		assert.Equal(t, weightPrefixName, r.Score)
		assert.Equal(t, "name", r.MatchField)
	}
}

func TestSearch_AliasExactMatch(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	results := idx.Search("cat", 0)
	require.NotEmpty(t, results)
	assert.Equal(t, "fs_read", results[0].Name)
	assert.Equal(t, weightExactAlias, results[0].Score)
	assert.Equal(t, "alias", results[0].MatchField)
}

func TestSearch_AliasPrefixMatch(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	// "ba" is a prefix of "bash" alias on exec_shell.
	results := idx.Search("ba", 0)
	require.NotEmpty(t, results)

	found := false
	for _, r := range results {
		if r.Name == "exec_shell" {
			found = true
			assert.Equal(t, weightPrefixAlias, r.Score)
			assert.Equal(t, "alias", r.MatchField)
			break
		}
	}
	assert.True(t, found, "exec_shell should match via alias prefix 'ba'")
}

func TestSearch_SearchHintMatch(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	results := idx.Search("terminal", 0)
	require.NotEmpty(t, results)
	assert.Equal(t, "exec_shell", results[0].Name)
	assert.Equal(t, weightSearchHint, results[0].Score)
	assert.Equal(t, "search_hint", results[0].MatchField)
}

func TestSearch_CategoryMatch(t *testing.T) {
	t.Parallel()

	// Use a tool whose name does NOT start with the category token,
	// so we isolate the category-matching path.
	c := New()
	c.RegisterCategory(Category{Name: "analytics"})
	c.Register("analytics", []*agent.Tool{
		{
			Name:        "metric_counter",
			Description: "counts metrics",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Exposure: agent.ExposureDefault,
				Category: "analytics",
			},
		},
	})

	idx := NewSearchIndex(c)
	results := idx.Search("analytics", 0)
	require.NotEmpty(t, results)
	assert.Equal(t, "metric_counter", results[0].Name)
	assert.Equal(t, weightCategory, results[0].Score)
	assert.Equal(t, "category", results[0].MatchField)
}

func TestSearch_DescriptionSubstring(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "misc"})
	c.Register("misc", []*agent.Tool{
		{
			Name:        "unique_tool",
			Description: "a special helper for analysis",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability:  agent.ToolCapability{Exposure: agent.ExposureDefault},
		},
	})

	idx := NewSearchIndex(c)
	results := idx.Search("analysis", 0)
	require.Len(t, results, 1)
	assert.Equal(t, "unique_tool", results[0].Name)
	assert.Equal(t, weightDescription, results[0].Score)
	assert.Equal(t, "description", results[0].MatchField)
}

func TestSearch_ActivityMatch(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	results := idx.Search("execute", 0)
	require.NotEmpty(t, results)

	found := false
	for _, r := range results {
		if r.Name == "exec_shell" && r.MatchField == "activity" {
			found = true
			assert.Equal(t, weightActivity, r.Score)
			break
		}
	}
	assert.True(t, found, "exec_shell should match via activity 'execute'")
}

func TestSearch_MultiTokenQuery(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	// "file read" — fs_read should score highest:
	// "file" matches search_hint (4.0) + "read" matches alias exact (7.0) = 11.0
	results := idx.Search("file read", 0)
	require.NotEmpty(t, results)
	assert.Equal(t, "fs_read", results[0].Name)
	// "file" via search_hint (4.0) + "read" via alias exact (7.0) = 11.0
	assert.Equal(t, weightSearchHint+weightExactAlias, results[0].Score)
}

func TestSearch_LimitParameter(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	// "file" matches multiple tools via search hints.
	allResults := idx.Search("file", 0)
	require.Greater(t, len(allResults), 1)

	limited := idx.Search("file", 1)
	require.Len(t, limited, 1)
	assert.Equal(t, allResults[0].Name, limited[0].Name)
}

func TestSearch_EmptyQuery(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	assert.Nil(t, idx.Search("", 0))
	assert.Nil(t, idx.Search("   ", 0))
}

func TestSearch_NoMatches(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	results := idx.Search("zzzznonexistent", 0)
	assert.Empty(t, results)
}

func TestSearch_CaseInsensitive(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	lower := idx.Search("fs_read", 0)
	upper := idx.Search("FS_READ", 0)
	mixed := idx.Search("Fs_Read", 0)

	require.NotEmpty(t, lower)
	require.Equal(t, len(lower), len(upper))
	require.Equal(t, len(lower), len(mixed))
	assert.Equal(t, lower[0].Name, upper[0].Name)
	assert.Equal(t, lower[0].Score, upper[0].Score)
}

func TestSearch_HiddenToolExcluded(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	// browser_internal is hidden — should not appear even if queried by name.
	results := idx.Search("browser_internal", 0)
	for _, r := range results {
		assert.NotEqual(t, "browser_internal", r.Name)
	}
}

func TestSearch_TiesBrokenByName(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "cat"})
	c.Register("cat", []*agent.Tool{
		{
			Name:        "z_tool",
			Description: "does something with data",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability:  agent.ToolCapability{Exposure: agent.ExposureDefault},
		},
		{
			Name:        "a_tool",
			Description: "does something with data",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability:  agent.ToolCapability{Exposure: agent.ExposureDefault},
		},
	})

	idx := NewSearchIndex(c)
	results := idx.Search("data", 0)
	require.Len(t, results, 2)
	assert.Equal(t, results[0].Score, results[1].Score)
	assert.Equal(t, "a_tool", results[0].Name, "ties broken by name ascending")
	assert.Equal(t, "z_tool", results[1].Name)
}

func TestSearch_Rebuild(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "cat"})
	c.Register("cat", []*agent.Tool{
		newTestToolWithCapability("tool_a", agent.ToolCapability{Exposure: agent.ExposureDefault}),
	})

	idx := NewSearchIndex(c)
	results := idx.Search("tool_a", 0)
	require.Len(t, results, 1)

	// Add a new tool and rebuild.
	c.Register("cat", []*agent.Tool{
		newTestToolWithCapability("tool_b", agent.ToolCapability{Exposure: agent.ExposureDefault}),
	})
	idx.Rebuild(c)

	results = idx.Search("tool_b", 0)
	require.Len(t, results, 1)
	assert.Equal(t, "tool_b", results[0].Name)
}

func TestSearch_NegativeLimit(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	allResults := idx.Search("file", 0)
	negResults := idx.Search("file", -1)
	assert.Equal(t, len(allResults), len(negResults))
}

func TestSearch_CategoryFieldPopulated(t *testing.T) {
	t.Parallel()

	c := buildTestCatalog()
	idx := NewSearchIndex(c)

	results := idx.Search("fs_read", 0)
	require.NotEmpty(t, results)
	assert.Equal(t, "filesystem", results[0].Category, "Category should be catalog-level category")
}

func BenchmarkSearch_200Entries(b *testing.B) {
	c := New()
	c.RegisterCategory(Category{Name: "bench", Enabled: true})

	tools := make([]*agent.Tool, 200)
	for i := range tools {
		n := strconv.Itoa(i)
		tools[i] = &agent.Tool{
			Name:        "tool_" + n,
			Description: "benchmark tool number " + n + " for testing search performance",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Aliases:     []string{"alias_" + n, "alt_" + n},
				Category:    "bench_cat_" + strconv.Itoa(i%5),
				SearchHints: []string{"hint_" + n, "keyword_" + n},
				Exposure:    agent.ExposureDefault,
				Activity:    agent.ActivityRead,
			},
		}
	}
	c.Register("bench", tools)

	idx := NewSearchIndex(c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.Search("tool_100 hint_50", 10)
	}
}
