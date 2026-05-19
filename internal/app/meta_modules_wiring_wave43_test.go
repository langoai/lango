package app

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/enttest"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/skill"
)

func TestWave43MetaLearningToolsPersistSearchStatsAndCleanupDryRun(t *testing.T) {
	ctx := context.Background()
	store := wave43KnowledgeStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, config.DefaultConfig(), receipts.NewStore())

	saveTool := findTool(tools, "save_learning")
	require.NotNil(t, saveTool)
	got, err := saveTool.Handler(ctx, map[string]interface{}{
		"trigger":       "wave43-tool",
		"error_pattern": "wave43 timeout",
		"diagnosis":     "retryable transient failure",
		"fix":           "retry with bounded backoff",
		"category":      string(entlearning.CategoryTimeout),
	})
	require.NoError(t, err)
	assert.Equal(t, "saved", got.(map[string]interface{})["status"])

	searchTool := findTool(tools, "search_learnings")
	require.NotNil(t, searchTool)
	got, err = searchTool.Handler(ctx, map[string]interface{}{
		"query":    "wave43",
		"category": string(entlearning.CategoryTimeout),
	})
	require.NoError(t, err)
	searchResult := got.(map[string]interface{})
	assert.Equal(t, 1, searchResult["count"])

	statsTool := findTool(tools, "learning_stats")
	require.NotNil(t, statsTool)
	got, err = statsTool.Handler(ctx, map[string]interface{}{})
	require.NoError(t, err)
	stats := got.(*knowledge.LearningStats)
	assert.Equal(t, 1, stats.TotalCount)
	assert.Equal(t, 1, stats.ByCategory[entlearning.CategoryTimeout])

	cleanupTool := findTool(tools, "learning_cleanup")
	require.NotNil(t, cleanupTool)
	got, err = cleanupTool.Handler(ctx, map[string]interface{}{
		"category": string(entlearning.CategoryTimeout),
		"dry_run":  true,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"would_delete": 1, "dry_run": true}, got)
}

func TestWave43MetaSkillToolsCreateFilterAndBlockTraversal(t *testing.T) {
	ctx := context.Background()
	store := wave43KnowledgeStore(t)
	skillsDir := t.TempDir()
	registry := wave43SkillRegistry(skillsDir)
	cfg := config.DefaultConfig()
	cfg.Modes = map[string]config.SessionMode{
		"wave43-mode": {
			Name:   "wave43-mode",
			Skills: []string{"wave43-skill"},
		},
	}
	tools := buildMetaTools(
		store,
		nil,
		registry,
		config.SkillConfig{SkillsDir: skillsDir},
		cfg,
		receipts.NewStore(),
	)

	createTool := findTool(tools, "create_skill")
	require.NotNil(t, createTool)
	for _, name := range []string{"wave43-skill", "wave43-hidden"} {
		got, err := createTool.Handler(ctx, map[string]interface{}{
			"name":        name,
			"description": "Wave43 deterministic skill",
			"type":        "composite",
			"definition":  `{"steps":["inspect","assert"]}`,
			"parameters":  `{"type":"object"}`,
		})
		require.NoError(t, err)
		assert.Equal(t, "active", got.(map[string]interface{})["status"])
	}

	listTool := findTool(tools, "list_skills")
	require.NotNil(t, listTool)
	got, err := listTool.Handler(session.WithModeName(ctx, "wave43-mode"), map[string]interface{}{
		"summary": true,
	})
	require.NoError(t, err)
	filtered := got.(map[string]interface{})
	assert.Equal(t, 1, filtered["count"])
	summaries := filtered["skills"].([]map[string]interface{})
	require.Len(t, summaries, 1)
	assert.Equal(t, "wave43-skill", summaries[0]["name"])
	assert.Equal(t, "Wave43 deterministic skill", summaries[0]["description"])

	viewTool := findTool(tools, "view_skill")
	require.NotNil(t, viewTool)
	got, err = viewTool.Handler(ctx, map[string]interface{}{
		"name": "wave43-skill",
	})
	require.NoError(t, err)
	view := got.(map[string]interface{})
	assert.Equal(t, "wave43-skill", view["name"])
	assert.Contains(t, view["content"], "Wave43 deterministic skill")
	wantSkillPath, err := filepath.EvalSymlinks(filepath.Join(skillsDir, "wave43-skill", "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, wantSkillPath, view["path"])

	got, err = viewTool.Handler(ctx, map[string]interface{}{
		"name": "wave43-skill",
		"path": "../wave43-hidden/SKILL.md",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "outside the skill directory")
}

func TestWave43AutomationBackgroundComponentLifecycleWithoutWorkers(t *testing.T) {
	t.Parallel()

	cfg := wave43ModuleConfig(t)
	cfg.Background.Enabled = true
	application := &App{Config: cfg}
	module := &automationModule{cfg: cfg, app: application, bus: eventbus.New()}

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{
			Store:        &stubSessionStore{},
			ReceiptStore: receipts.NewStore(),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Components, 1)

	component := result.Components[0].Component
	assert.Equal(t, "background-manager", component.Name())

	var wg sync.WaitGroup
	require.NoError(t, component.Start(context.Background(), &wg))
	require.NoError(t, component.Stop(context.Background()))
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "background").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "cron").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workflow").Enabled)
}

func TestWave43InitP2PDisabledOrMissingWalletKeepsFilesystemUntouched(t *testing.T) {
	t.Parallel()

	cfg := wave43ModuleConfig(t)
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "keys")
	cfg.P2P.ListenAddrs = []string{"not-a-real-multiaddr"}

	cfg.P2P.Enabled = false
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))
	assert.NoDirExists(t, cfg.P2P.KeyDir)

	cfg.P2P.Enabled = true
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))
	assert.NoDirExists(t, cfg.P2P.KeyDir)
}

func wave43KnowledgeStore(t *testing.T) *knowledge.Store {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "wave43-knowledge.db")
	client := enttest.Open(t, "sqlite3", "file:"+dbPath+"?_fk=1")
	t.Cleanup(func() { client.Close() })
	return knowledge.NewStore(client, zap.NewNop().Sugar())
}

func wave43SkillRegistry(skillsDir string) *skill.Registry {
	store := skill.NewFileSkillStore(skillsDir, zap.NewNop().Sugar())
	return skill.NewRegistry(store, nil, zap.NewNop().Sugar())
}

func wave43ModuleConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = ""
	cfg.Providers = nil
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "session.db")
	cfg.Skill.Enabled = false
	cfg.Tools.Browser.Enabled = false
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = false
	cfg.P2P.Workspace.Enabled = false
	cfg.Economy.Enabled = false
	cfg.SmartAccount.Enabled = false
	cfg.Cron.Enabled = false
	cfg.Background.Enabled = false
	cfg.Workflow.Enabled = false
	cfg.Knowledge.Enabled = false
	cfg.Graph.Enabled = false
	cfg.ObservationalMemory.Enabled = false
	cfg.AgentMemory.Enabled = false
	cfg.Librarian.Enabled = false
	cfg.Ontology.Enabled = false
	cfg.MCP.Enabled = false
	cfg.Observability.Enabled = false
	return cfg
}
