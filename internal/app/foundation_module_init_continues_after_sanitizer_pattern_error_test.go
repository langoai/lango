package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/receipts"
)

func TestFoundationModuleInitContinuesAfterSanitizerPatternError(t *testing.T) {
	t.Parallel()

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	cfg.Gatekeeper.CustomPatterns = []string{"["}
	cfg.Cron.Enabled = true
	cfg.Background.Enabled = true
	cfg.Workflow.Enabled = true

	result, err := (&foundationModule{cfg: cfg}).Init(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	fv, ok := result.Values[appinit.ProvidesSupervisor].(*foundationValues)
	require.True(t, ok)
	require.NotNil(t, fv)
	t.Cleanup(func() { require.NoError(t, fv.Store.Close()) })

	assert.Nil(t, fv.Sanitizer)
	assert.Equal(t, map[string]bool{
		"cron":       true,
		"background": true,
		"workflow":   true,
	}, fv.AutoAvail)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "browser").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "crypto").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "secrets").Enabled)
}

func TestIntelligenceModuleInitAgentMemoryCatalogWithoutStorage(t *testing.T) {
	t.Parallel()

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	cfg.AgentMemory.Enabled = true

	result, err := (&intelligenceModule{cfg: cfg, bus: eventbus.New()}).Init(
		context.Background(),
		staticResolver{
			appinit.ProvidesSupervisor: &foundationValues{
				Store:        &stubSessionStore{},
				ReceiptStore: receipts.NewStore(),
			},
			appinit.ProvidesBaseTools: []*agent.Tool{{Name: "foundationModuleInitContinuesAfterSanitizerPatternError_base"}},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, result)

	values, ok := result.Values[appinit.ProvidesKnowledge].(*intelligenceValues)
	require.True(t, ok)
	require.NotNil(t, values)
	assert.Nil(t, values.AgentMemoryStore)
	assert.Nil(t, values.KC)
	assert.Nil(t, values.MC)
	assert.Nil(t, values.GC)
	assert.Nil(t, values.LC)

	agentMemoryEntry := requireCatalogEntry(t, result.CatalogEntries, "agent_memory")
	assert.True(t, agentMemoryEntry.Enabled)
	assert.Equal(t, []string{
		"memory_agent_save",
		"memory_agent_recall",
		"memory_agent_forget",
	}, catalogEntryToolNames(agentMemoryEntry))
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "meta").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "graph").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "memory").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "librarian").Enabled)
}

func TestAutomationModuleInitAllOptionalSubsystemsDisabledCatalog(t *testing.T) {
	t.Parallel()

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	module := &automationModule{cfg: cfg, app: &App{Config: cfg}, bus: eventbus.New()}
	require.False(t, module.Enabled())

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{
			Store:        &stubSessionStore{},
			ReceiptStore: receipts.NewStore(),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	values, ok := result.Values[appinit.ProvidesAutomation].(*automationValues)
	require.True(t, ok)
	require.NotNil(t, values)
	assert.Nil(t, values.CronScheduler)
	assert.Nil(t, values.BackgroundManager)
	assert.Nil(t, values.WorkflowEngine)
	assert.NotNil(t, values.AgentRunStore)
	assert.Empty(t, result.Components)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "agent_control").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "task_tracking").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "cron").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "background").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workflow").Enabled)
}

func TestNetworkModuleInitFullyDisabledCatalogWithoutServices(t *testing.T) {
	t.Parallel()

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	module := &networkModule{cfg: cfg, bus: eventbus.New(), app: &App{ctx: context.Background()}}
	require.False(t, module.Enabled())

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{
			ReceiptStore: receipts.NewStore(),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Empty(t, result.Tools)
	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])
	assert.Nil(t, result.Values[appinit.ProvidesEconomy])
	assert.Nil(t, result.Values[appinit.ProvidesContract])
	assert.Nil(t, result.Values[appinit.ProvidesSmartAccount])
	assert.Nil(t, result.Values[appinit.ProvidesWorkspace])

	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "payment").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "contract").Enabled)
	p2pEntry := requireCatalogEntry(t, result.CatalogEntries, "p2p")
	assert.False(t, p2pEntry.Enabled)
	assert.NotContains(t, p2pEntry.Description, "payment required")
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "economy").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
	assert.Nil(t, foundationModuleInitContinuesAfterSanitizerPatternErrorCatalogEntry(result.CatalogEntries, "workspace"))
}

func TestExtensionModuleInitNoMCPServersAndIgnoredObservabilitySubfeatures(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { require.NoError(t, os.Chdir(origWD)) })
	t.Setenv("HOME", filepath.Join(tmp, "home"))

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	cfg.MCP.Enabled = true
	cfg.Observability.Enabled = false
	cfg.Observability.Health.Enabled = true
	cfg.Observability.Tokens.Enabled = true

	result, err := (&extensionModule{cfg: cfg, bus: eventbus.New()}).Init(context.Background(), staticResolver{})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Empty(t, result.Tools)
	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesMCP])
	assert.Nil(t, result.Values[appinit.ProvidesObservability])
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "mcp").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "observability").Enabled)
}

func foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = ""
	cfg.Providers = nil
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "session.db")
	cfg.Skill.Enabled = false
	cfg.Skill.SkillsDir = filepath.Join(t.TempDir(), "skills")
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
	cfg.Observability.Health.Enabled = false
	cfg.Observability.Tokens.Enabled = false
	cfg.Observability.Tokens.PersistHistory = false
	return cfg
}

func foundationModuleInitContinuesAfterSanitizerPatternErrorCatalogEntry(entries []appinit.CatalogEntry, category string) *appinit.CatalogEntry {
	for i := range entries {
		if entries[i].Category == category {
			return &entries[i]
		}
	}
	return nil
}
