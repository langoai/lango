package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/types"
)

func TestWave24FoundationModuleInitBuildsBaseToolsAndDisabledOptionalCatalog(t *testing.T) {
	t.Parallel()

	cfg := wave24ModuleConfig(t)
	mod := &foundationModule{cfg: cfg}

	result, err := mod.Init(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	fv, ok := result.Values[appinit.ProvidesSupervisor].(*foundationValues)
	require.True(t, ok)
	require.NotNil(t, fv)
	require.NotNil(t, fv.Supervisor)
	require.NotNil(t, fv.Store)
	t.Cleanup(func() {
		require.NoError(t, fv.Store.Close())
	})

	assert.Same(t, fv.Store, result.Values[appinit.ProvidesSessionStore])
	assert.Nil(t, fv.Crypto)
	assert.Nil(t, fv.Keys)
	assert.Nil(t, fv.Secrets)
	assert.Nil(t, fv.BrowserSM)
	assert.NotNil(t, fv.ReceiptStore)
	assert.NotNil(t, fv.Refs)
	assert.NotNil(t, fv.Scanner)
	assert.NotNil(t, fv.CmdGuard)
	assert.Equal(t, map[string]bool{
		"cron":       false,
		"background": false,
		"workflow":   false,
	}, fv.AutoAvail)

	baseTools, ok := result.Values[appinit.ProvidesBaseTools].([]*agent.Tool)
	require.True(t, ok)
	require.NotEmpty(t, baseTools)
	assert.NotEmpty(t, catalogEntryToolNames(requireCatalogEntry(t, result.CatalogEntries, "exec")))
	assert.NotEmpty(t, catalogEntryToolNames(requireCatalogEntry(t, result.CatalogEntries, "filesystem")))
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "exec").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "filesystem").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "web").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "browser").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "crypto").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "secrets").Enabled)
}

func TestWave24IntelligenceModuleInitCollectsUnavailableDependencyStatuses(t *testing.T) {
	t.Parallel()

	cfg := wave24ModuleConfig(t)
	cfg.Skill.Enabled = false
	cfg.Knowledge.Enabled = true
	cfg.ObservationalMemory.Enabled = true
	cfg.Librarian.Enabled = true
	mod := &intelligenceModule{cfg: cfg, bus: eventbus.New()}

	result, err := mod.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{
			Store:        &stubSessionStore{},
			ReceiptStore: receipts.NewStore(),
		},
		appinit.ProvidesBaseTools: []*agent.Tool{{Name: "wave24_base"}},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Tools)
	assert.Empty(t, result.Components)

	values, ok := result.Values[appinit.ProvidesKnowledge].(*intelligenceValues)
	require.True(t, ok)
	require.NotNil(t, values)
	assert.Nil(t, values.KC)
	assert.Nil(t, values.MC)
	assert.Nil(t, values.GC)
	assert.Nil(t, values.LC)
	assert.Nil(t, values.AB)
	assert.Nil(t, values.Observer)
	assert.Nil(t, values.SkillRegistry)
	assert.Nil(t, values.AgentMemoryStore)
	assert.Nil(t, values.OntologyBridge)

	statuses := wave24FeatureStatusesByName(values.FeatureStatuses)
	require.Contains(t, statuses, "Graph Store")
	require.Contains(t, statuses, "Knowledge")
	require.Contains(t, statuses, "Obs. Memory")
	require.Contains(t, statuses, "Librarian")

	assert.Equal(t, types.FeatureStatus{Name: "Graph Store", Enabled: false, Healthy: true}, statuses["Graph Store"])
	assert.False(t, statuses["Knowledge"].Enabled)
	assert.False(t, statuses["Knowledge"].Healthy)
	assert.Contains(t, statuses["Knowledge"].Reason, "requires EntStore")
	assert.False(t, statuses["Obs. Memory"].Enabled)
	assert.False(t, statuses["Obs. Memory"].Healthy)
	assert.Contains(t, statuses["Obs. Memory"].Reason, "requires EntStore")
	assert.False(t, statuses["Librarian"].Enabled)
	assert.False(t, statuses["Librarian"].Healthy)
	assert.Contains(t, statuses["Librarian"].Reason, "requires knowledge system")
	assert.Equal(t, 3, values.FeatureStatuses.SilentDisabledCount())

	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "meta").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "graph").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "memory").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "agent_memory").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "librarian").Enabled)
}

func TestWave24AutomationModuleInitKeepsControlPlaneWhenStoresAreUnavailable(t *testing.T) {
	t.Parallel()

	cfg := wave24ModuleConfig(t)
	cfg.Cron.Enabled = true
	cfg.Workflow.Enabled = true
	app := &App{Config: cfg}
	mod := &automationModule{cfg: cfg, app: app, bus: eventbus.New()}

	result, err := mod.Init(context.Background(), staticResolver{
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
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "background").Enabled)
	assert.Nil(t, wave24CatalogEntry(result.CatalogEntries, "cron"))
	assert.Nil(t, wave24CatalogEntry(result.CatalogEntries, "workflow"))
	assert.NotNil(t, findTool(result.Tools, "agent_spawn"))
	assert.NotNil(t, findTool(result.Tools, "task_create"))
}

func TestWave24NetworkModuleInitPaymentDependencyDisabledBranches(t *testing.T) {
	t.Parallel()

	cfg := wave24ModuleConfig(t)
	cfg.Payment.Enabled = true
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true
	cfg.Economy.Enabled = true
	cfg.SmartAccount.Enabled = true
	mod := &networkModule{cfg: cfg, bus: eventbus.New()}

	result, err := mod.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{},
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
	assert.Contains(t, p2pEntry.Description, "payment required")
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workspace").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
}

func TestWave24ExtensionModuleInitObservabilityEnabledWithoutPersistentStore(t *testing.T) {
	t.Parallel()

	cfg := wave24ModuleConfig(t)
	cfg.Observability.Enabled = true
	cfg.Observability.Health.Enabled = true
	cfg.Observability.Tokens.Enabled = true
	cfg.Observability.Tokens.PersistHistory = true
	mod := &extensionModule{cfg: cfg, bus: eventbus.New()}

	result, err := mod.Init(context.Background(), staticResolver{})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Tools)
	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesMCP])

	obsc, ok := result.Values[appinit.ProvidesObservability].(*observabilityComponents)
	require.True(t, ok)
	require.NotNil(t, obsc)
	assert.NotNil(t, obsc.collector)
	assert.NotNil(t, obsc.healthRegistry)
	assert.NotNil(t, obsc.tracker)
	assert.Nil(t, obsc.tokenStore)

	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "mcp").Enabled)
	assert.Nil(t, wave24CatalogEntry(result.CatalogEntries, "observability"))
}

func wave24ModuleConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = ""
	cfg.Providers = nil
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "session.db")
	cfg.Skill.SkillsDir = filepath.Join(t.TempDir(), "skills")
	cfg.Security.Signer.Provider = ""
	cfg.Tools.Browser.Enabled = false
	cfg.MCP.Enabled = false
	cfg.Observability.Enabled = false
	cfg.Observability.Health.Enabled = false
	cfg.Observability.Tokens.Enabled = false
	cfg.Observability.Tokens.PersistHistory = false
	return cfg
}

func wave24FeatureStatusesByName(collector *StatusCollector) map[string]types.FeatureStatus {
	if collector == nil {
		return nil
	}

	statuses := collector.All()
	out := make(map[string]types.FeatureStatus, len(statuses))
	for _, status := range statuses {
		out[status.Name] = status
	}
	return out
}

func wave24CatalogEntry(
	entries []appinit.CatalogEntry,
	category string,
) *appinit.CatalogEntry {
	for i := range entries {
		if entries[i].Category == category {
			return &entries[i]
		}
	}
	return nil
}
