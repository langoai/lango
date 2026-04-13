package toolcatalog

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

// echoHandler returns a handler that echoes its tool name and params.
func echoHandler(name string) agent.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{
			"tool":   name,
			"params": params,
		}, nil
	}
}

// buildIntegrationCatalog creates a realistic catalog with 24 tools across
// 6 categories with mixed exposure policies and safety levels.
func buildIntegrationCatalog() *Catalog {
	c := New()

	// --- Categories ---
	c.RegisterCategory(Category{Name: "filesystem", Description: "file system operations", ConfigKey: "tools.filesystem.enabled", Enabled: true})
	c.RegisterCategory(Category{Name: "exec", Description: "command execution", ConfigKey: "tools.exec.enabled", Enabled: true})
	c.RegisterCategory(Category{Name: "browser", Description: "web browser automation", ConfigKey: "tools.browser.enabled", Enabled: true})
	c.RegisterCategory(Category{Name: "crypto", Description: "cryptographic operations", ConfigKey: "tools.crypto.enabled", Enabled: true})
	c.RegisterCategory(Category{Name: "knowledge", Description: "knowledge management", ConfigKey: "tools.knowledge.enabled", Enabled: true})
	c.RegisterCategory(Category{Name: "internal", Description: "internal system tools", Enabled: true})

	// --- Filesystem tools (5 tools) ---
	c.Register("filesystem", []*agent.Tool{
		{Name: "fs_read", Description: "read a file from the filesystem", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Aliases: []string{"cat", "read"}, Category: "filesystem",
				SearchHints: []string{"file", "content", "open"},
				Exposure: agent.ExposureDefault, ReadOnly: true, Activity: agent.ActivityRead,
			}, Handler: echoHandler("fs_read")},
		{Name: "fs_write", Description: "write content to a file", SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Aliases: []string{"write", "save"}, Category: "filesystem",
				SearchHints: []string{"file", "create", "modify"},
				Exposure: agent.ExposureDefault, Activity: agent.ActivityWrite,
			}, Handler: echoHandler("fs_write")},
		{Name: "fs_list", Description: "list files in a directory", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Aliases: []string{"ls", "dir"}, Category: "filesystem",
				SearchHints: []string{"directory", "listing", "folder"},
				Exposure: agent.ExposureDeferred, ReadOnly: true, Activity: agent.ActivityRead,
			}, Handler: echoHandler("fs_list")},
		{Name: "fs_edit", Description: "edit a file in place", SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Aliases: []string{"edit", "sed"}, Category: "filesystem",
				SearchHints: []string{"modify", "replace", "patch"},
				Exposure: agent.ExposureDeferred, Activity: agent.ActivityWrite,
			}, Handler: echoHandler("fs_edit")},
		{Name: "fs_delete", Description: "delete a file or directory", SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Aliases: []string{"rm", "remove"}, Category: "filesystem",
				SearchHints: []string{"delete", "unlink"},
				Exposure: agent.ExposureDeferred, Activity: agent.ActivityWrite,
			}, Handler: echoHandler("fs_delete")},
	})

	// --- Exec tools (4 tools) ---
	c.Register("exec", []*agent.Tool{
		{Name: "exec_shell", Description: "execute a shell command", SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Aliases: []string{"shell", "bash", "sh"}, Category: "execution",
				SearchHints: []string{"command", "terminal", "run"},
				Exposure: agent.ExposureDefault, Activity: agent.ActivityExecute,
			}, Handler: echoHandler("exec_shell")},
		{Name: "exec_bg", Description: "run a command in background", SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Aliases: []string{"background"}, Category: "execution",
				SearchHints: []string{"async", "background", "detach"},
				Exposure: agent.ExposureDeferred, Activity: agent.ActivityExecute,
			}, Handler: echoHandler("exec_bg")},
		{Name: "exec_status", Description: "check status of a background job", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Aliases: []string{"jobs", "status"}, Category: "execution",
				SearchHints: []string{"job", "process", "running"},
				Exposure: agent.ExposureDeferred, ReadOnly: true, Activity: agent.ActivityQuery,
			}, Handler: echoHandler("exec_status")},
		{Name: "exec_stop", Description: "stop a background job", SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Aliases: []string{"kill", "stop"}, Category: "execution",
				SearchHints: []string{"terminate", "cancel"},
				Exposure: agent.ExposureDeferred, Activity: agent.ActivityManage,
			}, Handler: echoHandler("exec_stop")},
	})

	// --- Browser tools (4 tools) ---
	c.Register("browser", []*agent.Tool{
		{Name: "browser_navigate", Description: "navigate to a URL", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Aliases: []string{"goto", "open_url"}, Category: "web",
				SearchHints: []string{"url", "page", "navigate", "browse"},
				Exposure: agent.ExposureDefault, Activity: agent.ActivityRead,
			}, Handler: echoHandler("browser_navigate")},
		{Name: "browser_screenshot", Description: "capture a screenshot", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Aliases: []string{"screenshot", "capture"}, Category: "web",
				SearchHints: []string{"image", "screen", "visual"},
				Exposure: agent.ExposureDefault, ReadOnly: true, Activity: agent.ActivityRead,
			}, Handler: echoHandler("browser_screenshot")},
		{Name: "browser_action", Description: "perform a browser action (click, type, etc.)", SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Aliases: []string{"click", "type"}, Category: "web",
				SearchHints: []string{"interact", "form", "button"},
				Exposure: agent.ExposureDeferred, Activity: agent.ActivityExecute,
			}, Handler: echoHandler("browser_action")},
		{Name: "browser_internal_debug", Description: "internal browser debugging tool", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Exposure: agent.ExposureHidden, ReadOnly: true, Activity: agent.ActivityQuery,
			}, Handler: echoHandler("browser_internal_debug")},
	})

	// --- Crypto tools (4 tools) ---
	c.Register("crypto", []*agent.Tool{
		{Name: "crypto_hash", Description: "compute a cryptographic hash", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Aliases: []string{"hash", "digest"}, Category: "cryptography",
				SearchHints: []string{"sha256", "md5", "checksum"},
				Exposure: agent.ExposureDefault, ReadOnly: true, Activity: agent.ActivityRead,
			}, Handler: echoHandler("crypto_hash")},
		{Name: "crypto_encrypt", Description: "encrypt data", SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Aliases: []string{"encrypt"}, Category: "cryptography",
				SearchHints: []string{"aes", "cipher", "protect"},
				Exposure: agent.ExposureDeferred, Activity: agent.ActivityWrite,
			}, Handler: echoHandler("crypto_encrypt")},
		{Name: "crypto_sign", Description: "sign data with a private key", SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Aliases: []string{"sign"}, Category: "cryptography",
				SearchHints: []string{"signature", "verify", "key"},
				Exposure: agent.ExposureDeferred, Activity: agent.ActivityWrite,
			}, Handler: echoHandler("crypto_sign")},
		{Name: "crypto_keys", Description: "list available keys", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Aliases: []string{"keys", "keyring"}, Category: "cryptography",
				SearchHints: []string{"keypair", "public", "private"},
				Exposure: agent.ExposureDefault, ReadOnly: true, Activity: agent.ActivityQuery,
			}, Handler: echoHandler("crypto_keys")},
	})

	// --- Knowledge tools (4 tools) ---
	c.Register("knowledge", []*agent.Tool{
		{Name: "search_knowledge", Description: "search the knowledge base", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Aliases: []string{"search", "find"}, Category: "knowledge",
				SearchHints: []string{"query", "lookup", "retrieve"},
				Exposure: agent.ExposureDefault, ReadOnly: true, Activity: agent.ActivityQuery,
			}, Handler: echoHandler("search_knowledge")},
		{Name: "save_knowledge", Description: "save an item to the knowledge base", SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Aliases: []string{"remember", "store"}, Category: "knowledge",
				SearchHints: []string{"persist", "memorize"},
				Exposure: agent.ExposureDefault, Activity: agent.ActivityWrite,
			}, Handler: echoHandler("save_knowledge")},
		{Name: "knowledge_graph_query", Description: "query the knowledge graph", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Aliases: []string{"graph_query", "sparql"}, Category: "knowledge",
				SearchHints: []string{"graph", "triple", "relation"},
				Exposure: agent.ExposureDeferred, ReadOnly: true, Activity: agent.ActivityQuery,
			}, Handler: echoHandler("knowledge_graph_query")},
		{Name: "knowledge_import", Description: "import knowledge from external source", SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Aliases: []string{"import"}, Category: "knowledge",
				SearchHints: []string{"ingest", "load", "external"},
				Exposure: agent.ExposureDeferred, Activity: agent.ActivityWrite,
			}, Handler: echoHandler("knowledge_import")},
	})

	// --- Internal tools (3 tools, all hidden) ---
	c.Register("internal", []*agent.Tool{
		{Name: "internal_metrics", Description: "collect internal metrics", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Exposure: agent.ExposureHidden, ReadOnly: true, Activity: agent.ActivityQuery,
			}, Handler: echoHandler("internal_metrics")},
		{Name: "internal_debug", Description: "debug agent state", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Exposure: agent.ExposureHidden, ReadOnly: true, Activity: agent.ActivityQuery,
			}, Handler: echoHandler("internal_debug")},
		{Name: "internal_reset", Description: "reset internal state", SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Exposure: agent.ExposureHidden, Activity: agent.ActivityManage,
			}, Handler: echoHandler("internal_reset")},
	})

	return c
}

// TestIntegration_FullToolDiscoveryPipeline is an end-to-end test verifying
// the complete tool discovery pipeline: catalog registration, search index
// construction, dispatcher creation, and correct behavior of list/search/invoke.
func TestIntegration_FullToolDiscoveryPipeline(t *testing.T) {
	t.Parallel()

	catalog := buildIntegrationCatalog()
	index := NewSearchIndex(catalog)
	dispatcher := BuildDispatcher(catalog, index)

	require.Len(t, dispatcher, 4, "dispatcher should return 4 tools")

	listTool := dispatcher[0]
	invokeTool := dispatcher[1]
	searchTool := dispatcher[3]

	// --- Sub-tests ---

	t.Run("builtin_list returns only visible tools with deferred count hint", func(t *testing.T) {
		t.Parallel()

		result, err := listTool.Handler(context.Background(), map[string]interface{}{})
		require.NoError(t, err)

		m, ok := result.(map[string]interface{})
		require.True(t, ok)

		tools, ok := m["tools"].([]map[string]interface{})
		require.True(t, ok)

		// Count visible tools: ExposureDefault or ExposureAlwaysVisible only.
		// Visible: fs_read, fs_write, exec_shell, browser_navigate, browser_screenshot,
		//          crypto_hash, crypto_keys, search_knowledge, save_knowledge = 9
		assert.Len(t, tools, 9, "only visible tools should appear")

		// No hidden or deferred tool should appear.
		for _, tool := range tools {
			name := tool["name"].(string)
			assert.NotContains(t, name, "internal_", "hidden tools must not appear in list")
		}

		// Verify deferred count hint.
		deferredCount, ok := m["deferred_count"].(int)
		require.True(t, ok)
		assert.Equal(t, 11, deferredCount, "11 deferred tools should be counted")

		// Hint should be present because deferred tools exist.
		hint, ok := m["hint"].(string)
		require.True(t, ok)
		assert.Contains(t, hint, "builtin_search")
	})

	t.Run("builtin_search finds deferred tools by name", func(t *testing.T) {
		t.Parallel()

		result, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"query": "fs_list",
		})
		require.NoError(t, err)

		m := result.(map[string]interface{})
		results := m["results"].([]map[string]interface{})
		require.NotEmpty(t, results)
		assert.Equal(t, "fs_list", results[0]["name"])
	})

	t.Run("builtin_search finds deferred tools by alias", func(t *testing.T) {
		t.Parallel()

		result, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"query": "ls",
		})
		require.NoError(t, err)

		m := result.(map[string]interface{})
		results := m["results"].([]map[string]interface{})
		require.NotEmpty(t, results)
		assert.Equal(t, "fs_list", results[0]["name"])
	})

	t.Run("builtin_search finds deferred tools by hint", func(t *testing.T) {
		t.Parallel()

		result, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"query": "directory",
		})
		require.NoError(t, err)

		m := result.(map[string]interface{})
		results := m["results"].([]map[string]interface{})
		require.NotEmpty(t, results)

		found := false
		for _, r := range results {
			if r["name"] == "fs_list" {
				found = true
				break
			}
		}
		assert.True(t, found, "fs_list should be discoverable by search hint 'directory'")
	})

	t.Run("builtin_search does NOT find hidden tools", func(t *testing.T) {
		t.Parallel()

		hiddenNames := []string{"browser_internal_debug", "internal_metrics", "internal_debug", "internal_reset"}

		for _, hidden := range hiddenNames {
			result, err := searchTool.Handler(context.Background(), map[string]interface{}{
				"query": hidden,
			})
			require.NoError(t, err)

			m := result.(map[string]interface{})
			results := m["results"].([]map[string]interface{})
			for _, r := range results {
				assert.NotEqual(t, hidden, r["name"],
					"hidden tool %q must not appear in search results", hidden)
			}
		}
	})

	t.Run("builtin_invoke executes visible safe tools", func(t *testing.T) {
		t.Parallel()

		result, err := invokeTool.Handler(context.Background(), map[string]interface{}{
			"tool_name": "browser_navigate",
			"params":    map[string]interface{}{"url": "https://example.com"},
		})
		require.NoError(t, err)

		m := result.(map[string]interface{})
		assert.Equal(t, "browser_navigate", m["tool"])

		inner := m["result"].(map[string]interface{})
		assert.Equal(t, "browser_navigate", inner["tool"])
	})

	t.Run("builtin_invoke executes deferred safe tools", func(t *testing.T) {
		t.Parallel()

		result, err := invokeTool.Handler(context.Background(), map[string]interface{}{
			"tool_name": "fs_list",
			"params":    map[string]interface{}{"path": "/tmp"},
		})
		require.NoError(t, err)

		m := result.(map[string]interface{})
		assert.Equal(t, "fs_list", m["tool"])

		inner := m["result"].(map[string]interface{})
		assert.Equal(t, "fs_list", inner["tool"])
	})

	t.Run("builtin_invoke blocks dangerous tools regardless of exposure", func(t *testing.T) {
		t.Parallel()

		dangerousTools := []struct {
			name     string
			exposure string
		}{
			{name: "fs_write", exposure: "visible"},
			{name: "exec_shell", exposure: "visible"},
			{name: "fs_edit", exposure: "deferred"},
			{name: "fs_delete", exposure: "deferred"},
			{name: "exec_bg", exposure: "deferred"},
			{name: "crypto_encrypt", exposure: "deferred"},
		}

		for _, tt := range dangerousTools {
			t.Run(fmt.Sprintf("%s_%s", tt.name, tt.exposure), func(t *testing.T) {
				t.Parallel()

				_, err := invokeTool.Handler(context.Background(), map[string]interface{}{
					"tool_name": tt.name,
				})
				require.Error(t, err)
				assert.Contains(t, err.Error(), "requires approval",
					"dangerous tool %q (%s) must be blocked", tt.name, tt.exposure)
			})
		}
	})
}

func TestIntegration_ListVisibleTools(t *testing.T) {
	t.Parallel()

	catalog := buildIntegrationCatalog()

	t.Run("all visible tools across categories", func(t *testing.T) {
		t.Parallel()

		schemas := catalog.ListVisibleTools("")
		var names []string
		for _, s := range schemas {
			names = append(names, s.Name)
		}

		// All ExposureDefault tools (no ExposureAlwaysVisible in this catalog).
		wantVisible := []string{
			"fs_read", "fs_write",
			"exec_shell",
			"browser_navigate", "browser_screenshot",
			"crypto_hash", "crypto_keys",
			"search_knowledge", "save_knowledge",
		}
		assert.Equal(t, wantVisible, names)

		// Deferred and hidden tools must not be present.
		for _, name := range names {
			assert.NotContains(t, name, "internal_")
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		t.Parallel()

		schemas := catalog.ListVisibleTools("crypto")
		var names []string
		for _, s := range schemas {
			names = append(names, s.Name)
		}
		assert.Equal(t, []string{"crypto_hash", "crypto_keys"}, names)
	})
}

func TestIntegration_SearchableEntries(t *testing.T) {
	t.Parallel()

	catalog := buildIntegrationCatalog()
	entries := catalog.SearchableEntries()

	var names []string
	for _, e := range entries {
		names = append(names, e.Tool.Name)
	}

	// Searchable = everything except hidden tools.
	// Hidden: browser_internal_debug, internal_metrics, internal_debug, internal_reset
	assert.Len(t, entries, 20, "24 total - 4 hidden = 20 searchable")

	for _, name := range names {
		assert.NotEqual(t, "browser_internal_debug", name)
		assert.NotEqual(t, "internal_metrics", name)
		assert.NotEqual(t, "internal_debug", name)
		assert.NotEqual(t, "internal_reset", name)
	}
}

func TestIntegration_DeferredToolCount(t *testing.T) {
	t.Parallel()

	catalog := buildIntegrationCatalog()

	// Deferred tools: fs_list, fs_edit, fs_delete, exec_bg, exec_status, exec_stop,
	//                 browser_action, crypto_encrypt, crypto_sign,
	//                 knowledge_graph_query, knowledge_import = 11
	assert.Equal(t, 11, catalog.DeferredToolCount())
}

func TestIntegration_SearchRanking(t *testing.T) {
	t.Parallel()

	catalog := buildIntegrationCatalog()
	index := NewSearchIndex(catalog)

	t.Run("exact name outranks alias", func(t *testing.T) {
		t.Parallel()

		results := index.Search("fs_read", 0)
		require.NotEmpty(t, results)

		// fs_read should be first with exact name score.
		assert.Equal(t, "fs_read", results[0].Name)
		assert.Equal(t, weightExactName, results[0].Score)
		assert.Equal(t, "name", results[0].MatchField)
	})

	t.Run("exact alias outranks description", func(t *testing.T) {
		t.Parallel()

		results := index.Search("cat", 0)
		require.NotEmpty(t, results)
		assert.Equal(t, "fs_read", results[0].Name)
		assert.Equal(t, weightExactAlias, results[0].Score)
		assert.Equal(t, "alias", results[0].MatchField)
	})

	t.Run("search hint outranks description-only match", func(t *testing.T) {
		t.Parallel()

		results := index.Search("terminal", 0)
		require.NotEmpty(t, results)
		assert.Equal(t, "exec_shell", results[0].Name)
		assert.Equal(t, weightSearchHint, results[0].Score)
		assert.Equal(t, "search_hint", results[0].MatchField)
	})

	t.Run("multi-token scoring is additive", func(t *testing.T) {
		t.Parallel()

		// "file read" for fs_read:
		// "file" -> search_hint (4.0), "read" -> exact alias (7.0) = 11.0
		results := index.Search("file read", 0)
		require.NotEmpty(t, results)
		assert.Equal(t, "fs_read", results[0].Name)
		assert.Equal(t, weightSearchHint+weightExactAlias, results[0].Score)
	})

	t.Run("name prefix ranks above alias exact", func(t *testing.T) {
		t.Parallel()

		// "fs_" matches 5 tools as name prefix (score 8 each).
		results := index.Search("fs_", 0)
		require.GreaterOrEqual(t, len(results), 5)

		for _, r := range results[:5] {
			assert.Contains(t, r.Name, "fs_")
			assert.Equal(t, weightPrefixName, r.Score)
		}
	})

	t.Run("ties broken by name ascending", func(t *testing.T) {
		t.Parallel()

		// "file" matches fs_read, fs_write, fs_list via search hints (all score 4).
		results := index.Search("file", 0)
		require.GreaterOrEqual(t, len(results), 3)

		// Among equal-scored results, names should be ascending.
		for i := 1; i < len(results); i++ {
			if results[i].Score == results[i-1].Score {
				assert.True(t, results[i].Name > results[i-1].Name,
					"equal scores should be tie-broken by name ascending: %q vs %q",
					results[i-1].Name, results[i].Name)
			}
		}
	})
}

func TestIntegration_ToolCountAndCategories(t *testing.T) {
	t.Parallel()

	catalog := buildIntegrationCatalog()

	assert.Equal(t, 24, catalog.ToolCount(), "total tool count")

	cats := catalog.ListCategories()
	assert.Len(t, cats, 6)

	// Sorted by name.
	assert.Equal(t, "browser", cats[0].Name)
	assert.Equal(t, "crypto", cats[1].Name)
	assert.Equal(t, "exec", cats[2].Name)
	assert.Equal(t, "filesystem", cats[3].Name)
	assert.Equal(t, "internal", cats[4].Name)
	assert.Equal(t, "knowledge", cats[5].Name)
}

func TestIntegration_ExposurePolicyCounts(t *testing.T) {
	t.Parallel()

	catalog := buildIntegrationCatalog()

	// Count by exposure type across all tools.
	allTools := catalog.ListTools("")
	var visibleCount, deferredCount, hiddenCount int
	for _, s := range allTools {
		switch s.Exposure {
		case "":
			visibleCount++ // ExposureDefault omits the exposure field
		case "deferred":
			deferredCount++
		case "hidden":
			hiddenCount++
		}
	}

	assert.Equal(t, 9, visibleCount, "visible (default) tool count")
	assert.Equal(t, 11, deferredCount, "deferred tool count")
	assert.Equal(t, 4, hiddenCount, "hidden tool count")
}

func TestIntegration_BuiltinListNoHintWhenNoDeferred(t *testing.T) {
	t.Parallel()

	c := New()
	c.RegisterCategory(Category{Name: "simple", Enabled: true})
	c.Register("simple", []*agent.Tool{
		{Name: "tool_a", Description: "simple tool A", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{Exposure: agent.ExposureDefault},
			Handler:    echoHandler("tool_a")},
		{Name: "tool_b", Description: "simple tool B", SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{Exposure: agent.ExposureDefault},
			Handler:    echoHandler("tool_b")},
	})

	idx := NewSearchIndex(c)
	dispatcher := BuildDispatcher(c, idx)
	listTool := dispatcher[0]

	result, err := listTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, 0, m["deferred_count"])
	_, hasHint := m["hint"]
	assert.False(t, hasHint, "no hint when no deferred tools exist")
}
