package app

import (
	"context"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
	"github.com/langoai/lango/internal/toolcatalog"
)

// ─── Layer 1: Helper Unit Tests ───

func TestBuildCatalogFromEntries_Basic(t *testing.T) {
	t.Parallel()

	toolA := &agent.Tool{Name: "tool_a"}
	toolB := &agent.Tool{Name: "tool_b"}
	toolC := &agent.Tool{Name: "tool_c"}

	entries := []appinit.CatalogEntry{
		{Category: "alpha", Description: "Alpha tools", Enabled: true, Tools: []*agent.Tool{toolA, toolB}},
		{Category: "beta", Description: "Beta tools", Enabled: true, Tools: []*agent.Tool{toolC}},
		{Category: "gamma", Description: "Gamma (disabled)", Enabled: false},
	}

	catalog := buildCatalogFromEntries(entries)

	// Total tool count: 3 (only enabled entries with tools contribute).
	assert.Equal(t, 3, catalog.ToolCount())

	// Category names.
	categories := catalog.ListCategories()
	catNames := make([]string, len(categories))
	for i, c := range categories {
		catNames[i] = c.Name
	}
	sort.Strings(catNames)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, catNames)

	// Enabled / disabled status.
	catMap := make(map[string]bool, len(categories))
	for _, c := range categories {
		catMap[c.Name] = c.Enabled
	}
	assert.True(t, catMap["alpha"])
	assert.True(t, catMap["beta"])
	assert.False(t, catMap["gamma"])

	// Tool names per category.
	assert.Equal(t, []string{"tool_a", "tool_b"}, catalog.ToolNamesForCategory("alpha"))
	assert.Equal(t, []string{"tool_c"}, catalog.ToolNamesForCategory("beta"))
	assert.Empty(t, catalog.ToolNamesForCategory("gamma"))
}

func TestBuildCatalogFromEntries_DuplicateCategory(t *testing.T) {
	t.Parallel()

	toolA := &agent.Tool{Name: "mcp_server1_tool"}
	toolB := &agent.Tool{Name: "mcp_server2_tool"}

	entries := []appinit.CatalogEntry{
		{Category: "mcp", Description: "MCP tools (server 1)", Enabled: true, Tools: []*agent.Tool{toolA}},
		{Category: "mcp", Description: "MCP management tools", Enabled: true, Tools: []*agent.Tool{toolB}},
	}

	catalog := buildCatalogFromEntries(entries)

	// Both tools should be registered under "mcp".
	assert.Equal(t, 2, catalog.ToolCount())
	names := catalog.ToolNamesForCategory("mcp")
	assert.Equal(t, []string{"mcp_server1_tool", "mcp_server2_tool"}, names)
}

func TestRegisterPostBuildLifecycle_Names(t *testing.T) {
	t.Parallel()

	t.Run("no channels", func(t *testing.T) {
		t.Parallel()

		reg := lifecycle.NewRegistry()
		a := &App{
			registry: reg,
		}
		// Gateway is set by the caller of registerPostBuildLifecycle;
		// we just need a non-nil Gateway to avoid panic.
		a.Gateway = initGateway(config.DefaultConfig(), nil, nil, nil)

		registerPostBuildLifecycle(a)
		assert.Equal(t, []string{"gateway"}, reg.Names())
	})

	t.Run("with channels", func(t *testing.T) {
		t.Parallel()

		reg := lifecycle.NewRegistry()
		a := &App{
			registry: reg,
			Channels: []Channel{&noopChannel{}, &noopChannel{}},
		}
		a.Gateway = initGateway(config.DefaultConfig(), nil, nil, nil)

		registerPostBuildLifecycle(a)
		assert.Equal(t, []string{"gateway", "channel-0", "channel-1"}, reg.Names())
	})
}

// noopChannel satisfies the Channel interface for testing.
type noopChannel struct{}

func (n *noopChannel) Name() string                  { return "noop" }
func (n *noopChannel) Start(_ context.Context) error { return nil }
func (n *noopChannel) Stop(_ context.Context) error  { return nil }

// ─── Layer 2: Integration Parity Tests ───

func TestAppNew_DefaultConfig_Parity(t *testing.T) {
	testutil.SkipShort(t)

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	cfg.Session.DatabasePath = filepath.Join(cfg.DataRoot, "test.db")
	// Ensure no provider is set so it doesn't try to init external API clients.
	cfg.Agent.Provider = ""

	client := testutil.TestEntClient(t)
	boot := &bootstrap.Result{
		Config:      cfg,
		Storage:     storage.NewFacade(nil, nil, storage.WithEntClient(client)),
		ProfileName: "test",
	}

	application, err := New(boot)
	require.NoError(t, err, "app.New() must succeed with default config")
	t.Cleanup(func() {
		application.cancel()
	})

	catalog := application.ToolCatalog
	require.NotNil(t, catalog)

	// 1. Enabled categories.
	categories := catalog.ListCategories()
	enabledNames := make([]string, 0)
	disabledNames := make([]string, 0)
	for _, c := range categories {
		if c.Enabled {
			enabledNames = append(enabledNames, c.Name)
		} else {
			disabledNames = append(disabledNames, c.Name)
		}
	}
	sort.Strings(enabledNames)

	assert.Contains(t, enabledNames, "exec")
	assert.Contains(t, enabledNames, "filesystem")
	assert.Contains(t, enabledNames, "output")

	// 2. Disabled categories.
	for _, name := range []string{"browser", "crypto", "secrets", "meta", "graph", "memory", "agent_memory", "librarian", "mcp", "observability"} {
		assert.Contains(t, disabledNames, name, "expected %q to be disabled", name)
	}

	// 3. Tool count: exec (4) + filesystem (7) + output + dispatcher = at least 11.
	assert.GreaterOrEqual(t, catalog.ToolCount(), 11,
		"default config should register at least 11 tools (exec+filesystem)")

	// 4. Dispatcher tools: verify BuildDispatcher produces the expected 4 tools.
	// These are appended to the tool list passed to the agent at B3.
	searchIndex := toolcatalog.NewSearchIndex(catalog)
	dispatcherTools := toolcatalog.BuildDispatcher(catalog, searchIndex)
	require.Len(t, dispatcherTools, 4)
	dispatcherNames := make(map[string]bool, 4)
	for _, dt := range dispatcherTools {
		dispatcherNames[dt.Name] = true
	}
	assert.True(t, dispatcherNames["builtin_list"], "builtin_list dispatcher tool missing")
	assert.True(t, dispatcherNames["builtin_invoke"], "builtin_invoke dispatcher tool missing")
	assert.True(t, dispatcherNames["builtin_health"], "builtin_health dispatcher tool missing")
	assert.True(t, dispatcherNames["builtin_search"], "builtin_search dispatcher tool missing")

	// 5. Lifecycle names include "gateway".
	regNames := application.registry.Names()
	assert.Contains(t, regNames, "gateway")

	// 6. Lifecycle names do NOT include disabled components.
	for _, absent := range []string{"p2p-node", "cron-scheduler", "mcp-manager", "channel-0"} {
		assert.NotContains(t, regNames, absent, "expected %q to be absent from lifecycle", absent)
	}

	// 7. Non-nil core fields.
	assert.NotNil(t, application.Store)
	assert.NotNil(t, application.Gateway)
	assert.NotNil(t, application.ToolCatalog)
	assert.NotNil(t, application.Agent)

	// 8. Nil optional fields (disabled features).
	assert.Nil(t, application.P2PNode)
	assert.Nil(t, application.CronScheduler)
	assert.Nil(t, application.MCPManager)
	assert.Nil(t, application.KnowledgeStore)
}

func TestAppNew_FeaturesEnabled_Parity(t *testing.T) {
	testutil.SkipShort(t)

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	cfg.Session.DatabasePath = filepath.Join(cfg.DataRoot, "test.db")
	cfg.Agent.Provider = ""

	// Enable specific features.
	cfg.Knowledge.Enabled = true
	cfg.Graph.Enabled = true
	cfg.ObservationalMemory.Enabled = true
	cfg.Cron.Enabled = true
	// Intentionally keep these disabled for test stability:
	// - Embedding.Provider = "" (no embedding/RAG)
	// - Security.Signer.Provider = "" (no crypto tools)
	// - MCP.Enabled = false
	// - P2P.Enabled = false
	// - Payment.Enabled = false

	client := testutil.TestEntClient(t)
	boot := &bootstrap.Result{
		Config:      cfg,
		Storage:     storage.NewFacade(nil, nil, storage.WithEntClient(client)),
		ProfileName: "test",
	}

	application, err := New(boot)
	require.NoError(t, err, "app.New() must succeed with features enabled")
	t.Cleanup(func() {
		application.cancel()
	})

	catalog := application.ToolCatalog
	require.NotNil(t, catalog)

	// 1. Additional enabled categories.
	categories := catalog.ListCategories()
	enabledNames := make([]string, 0)
	disabledNames := make([]string, 0)
	for _, c := range categories {
		if c.Enabled {
			enabledNames = append(enabledNames, c.Name)
		} else {
			disabledNames = append(disabledNames, c.Name)
		}
	}

	assert.Contains(t, enabledNames, "meta", "knowledge should enable meta category")
	assert.Contains(t, enabledNames, "graph", "graph should be enabled")
	assert.Contains(t, enabledNames, "memory", "observational memory should be enabled")
	assert.Contains(t, enabledNames, "cron", "cron should be enabled")

	// 2. Background/workflow still disabled.
	assert.Contains(t, disabledNames, "background")
	assert.Contains(t, disabledNames, "workflow")

	// 3. Lifecycle names include feature-specific components.
	regNames := application.registry.Names()
	assert.Contains(t, regNames, "memory-buffer")
	assert.Contains(t, regNames, "graph-buffer")
	assert.Contains(t, regNames, "cron-scheduler")

	// 4. Non-nil feature fields.
	assert.NotNil(t, application.KnowledgeStore, "KnowledgeStore should be set")
	assert.NotNil(t, application.MemoryStore, "MemoryStore should be set")
	assert.NotNil(t, application.GraphStore, "GraphStore should be set")
	assert.NotNil(t, application.CronScheduler, "CronScheduler should be set")

	// 5. Still nil (disabled features).
	assert.Nil(t, application.P2PNode)
	assert.Nil(t, application.MCPManager)
}
