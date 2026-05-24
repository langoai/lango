package app

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/tokenusage"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
)

func TestIntelligenceModuleInitWiresGraphAdmissionWithoutExternalServices(t *testing.T) {
	t.Parallel()

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = filepath.Join(t.TempDir(), "graph.db")
	cfg.Ontology.Governance.AdmissionMode = config.OntologyAdmissionModeObserve
	module := &intelligenceModule{cfg: cfg, bus: eventbus.New()}

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{
			Store:        &stubSessionStore{},
			ReceiptStore: receipts.NewStore(),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	values, ok := result.Values[appinit.ProvidesKnowledge].(*intelligenceValues)
	require.True(t, ok)
	require.NotNil(t, values)
	require.NotNil(t, values.GC)
	t.Cleanup(func() { require.NoError(t, values.GC.store.Close()) })

	assert.NotNil(t, values.GC.buffer)
	assert.NotNil(t, values.GC.admissionPolicy)
	assert.Same(t, values.GC, result.Values[appinit.ProvidesGraph])
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "graph").Enabled)
	assert.NotEmpty(t, catalogEntryToolNames(requireCatalogEntry(t, result.CatalogEntries, "graph")))
	assert.NotEmpty(t, result.Components)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "meta").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "memory").Enabled)
}

func TestIntelligenceModuleInitWiresLocalStoresAndOntologyWithoutProviders(t *testing.T) {
	t.Parallel()

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	cfg.Knowledge.Enabled = true
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = filepath.Join(t.TempDir(), "graph.db")
	cfg.ObservationalMemory.Enabled = true
	cfg.Librarian.Enabled = true
	cfg.AgentMemory.Enabled = true
	cfg.Ontology.Enabled = true
	cfg.Ontology.Governance.Enabled = true
	cfg.Ontology.Governance.AdmissionMode = config.OntologyAdmissionModeObserve
	cfg.Ontology.Exchange.Enabled = true

	client := testutil.TestEntClient(t)
	store := session.NewEntStoreWithClient(client)
	facade := storage.NewFacade(nil, nil, storage.WithEntClient(client))
	module := &intelligenceModule{
		cfg:  cfg,
		boot: &bootstrap.Result{Storage: facade},
		bus:  eventbus.New(),
	}

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{
			Store:        store,
			ReceiptStore: nil,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	values, ok := result.Values[appinit.ProvidesKnowledge].(*intelligenceValues)
	require.True(t, ok)
	require.NotNil(t, values)
	require.NotNil(t, values.GC)
	t.Cleanup(func() { require.NoError(t, values.GC.store.Close()) })

	assert.NotNil(t, values.KC)
	assert.NotNil(t, values.MC)
	assert.NotNil(t, values.LC)
	assert.NotNil(t, values.AgentMemoryStore)
	assert.NotNil(t, values.OntologyBridge)
	assert.Same(t, values.GC, result.Values[appinit.ProvidesGraph])
	assert.Same(t, values.MC, result.Values[appinit.ProvidesMemory])
	assert.Same(t, values.LC, result.Values[appinit.ProvidesLibrarian])
	assert.Same(t, values.SkillRegistry, result.Values[appinit.ProvidesSkills])
	require.NotNil(t, values.FeatureStatuses)
	assert.GreaterOrEqual(t, values.FeatureStatuses.SilentDisabledCount(), 0)

	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "ontology").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "meta").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "graph").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "memory").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "agent_memory").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "librarian").Enabled)
	assert.NotNil(t, findTool(result.Tools, "ontology_list_types"))
	assert.NotNil(t, findTool(result.Tools, "librarian_pending_inquiries"))
	assert.NotNil(t, findTool(result.Tools, "memory_agent_save"))
	assert.Contains(t, modulesInitDeterministicBranchesComponentNames(result.Components), "librarian-proactive-buffer")
}

func TestNetworkModuleInitUsesBootStorageFacadeWhenPaymentIsDisabled(t *testing.T) {
	t.Parallel()

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true
	module := &networkModule{
		cfg:  cfg,
		boot: &bootstrap.Result{Storage: storage.NewFacade(nil, nil)},
		bus:  eventbus.New(),
		app:  &App{ctx: context.Background()},
	}

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
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workspace").Enabled)
	p2pEntry := requireCatalogEntry(t, result.CatalogEntries, "p2p")
	assert.False(t, p2pEntry.Enabled)
	assert.Contains(t, p2pEntry.Description, "payment required")
}

func TestNetworkModuleInitRegistersEconomyToolsWhenPaymentIsDisabled(t *testing.T) {
	t.Parallel()

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = false
	cfg.Economy.Enabled = true
	cfg.Economy.Pricing.Enabled = true
	cfg.Economy.Negotiate.Enabled = true
	cfg.Economy.Escrow.Enabled = true
	module := &networkModule{
		cfg: cfg,
		bus: eventbus.New(),
		app: &App{ctx: context.Background()},
	}

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{
			ReceiptStore: receipts.NewStore(),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	values, ok := result.Values[appinit.ProvidesEconomy].(*economyComponents)
	require.True(t, ok)
	require.NotNil(t, values)
	t.Cleanup(func() {
		if values.sentinelEngine != nil {
			require.NoError(t, values.sentinelEngine.Stop())
		}
	})

	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])
	assert.Nil(t, result.Values[appinit.ProvidesContract])
	assert.Nil(t, result.Values[appinit.ProvidesSmartAccount])
	assert.NotNil(t, values.budgetEngine)
	assert.NotNil(t, values.riskEngine)
	assert.NotNil(t, values.pricingEngine)
	assert.NotNil(t, values.negotiationEngine)
	assert.NotNil(t, values.escrowEngine)
	assert.NotNil(t, values.escrowSettler)
	assert.NotNil(t, values.sentinelEngine)

	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "economy").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "escrow").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "sentinel").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "payment").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "contract").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "p2p").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
	assert.NotNil(t, findTool(result.Tools, "economy_budget_allocate"))
	assert.NotNil(t, findTool(result.Tools, "economy_price_quote"))
	assert.NotNil(t, findTool(result.Tools, "economy_negotiate"))
	assert.NotNil(t, findTool(result.Tools, "economy_escrow_create"))
	assert.NotNil(t, findTool(result.Tools, "sentinel_status"))
}

func TestExtensionModuleInitRegistersMCPManagementForDisabledConfiguredServer(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { require.NoError(t, os.Chdir(origWD)) })
	t.Setenv("HOME", filepath.Join(tmp, "home"))

	disabled := false
	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	cfg.MCP.Enabled = true
	cfg.MCP.Servers = map[string]config.MCPServerConfig{
		"disabled-fixture": {
			Enabled: &disabled,
			Command: "not-started",
		},
	}
	module := &extensionModule{cfg: cfg, bus: eventbus.New()}

	result, err := module.Init(context.Background(), staticResolver{})
	require.NoError(t, err)
	require.NotNil(t, result)

	mcpc, ok := result.Values[appinit.ProvidesMCP].(*mcpComponents)
	require.True(t, ok)
	require.NotNil(t, mcpc)
	require.NotNil(t, mcpc.manager)
	assert.Empty(t, mcpc.tools)
	assert.Len(t, result.Components, 1)
	assert.Nil(t, result.Values[appinit.ProvidesObservability])

	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "mcp").Enabled)
	assert.NotNil(t, findTool(result.Tools, "mcp_status"))
	assert.NotNil(t, findTool(result.Tools, "mcp_tools"))
	require.NoError(t, result.Components[0].Component.Start(context.Background(), &sync.WaitGroup{}))
	require.NoError(t, result.Components[0].Component.Stop(context.Background()))

	statusTool := findTool(result.Tools, "mcp_status")
	status, err := statusTool.Handler(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "No MCP servers configured.", status)
}

func TestExtensionModuleInitAddsObservabilityTokenCleanupLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := testutil.TestEntClient(t)
	facade := storage.NewFacade(nil, nil, storage.WithEntClient(client))
	tokenStore := facade.TokenStore()
	now := time.Now().UTC()
	require.NoError(t, tokenStore.Save(observability.TokenUsage{
		Provider:     "provider",
		Model:        "model",
		SessionKey:   "modulesInitDeterministicBranches9-old",
		AgentName:    "agent",
		InputTokens:  1,
		OutputTokens: 2,
		TotalTokens:  3,
		Timestamp:    now.AddDate(0, 0, -10),
	}))
	require.NoError(t, tokenStore.Save(observability.TokenUsage{
		Provider:     "provider",
		Model:        "model",
		SessionKey:   "modulesInitDeterministicBranches9-recent",
		AgentName:    "agent",
		InputTokens:  4,
		OutputTokens: 5,
		TotalTokens:  9,
		Timestamp:    now.AddDate(0, 0, -1),
	}))

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	cfg.Observability.Enabled = true
	cfg.Observability.Tokens.Enabled = true
	cfg.Observability.Tokens.PersistHistory = true
	cfg.Observability.Tokens.RetentionDays = 3
	module := &extensionModule{
		cfg:  cfg,
		boot: &bootstrap.Result{Storage: facade},
		bus:  eventbus.New(),
	}

	result, err := module.Init(ctx, staticResolver{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Components, 1)
	assert.Equal(t, "observability-token-cleanup", result.Components[0].Component.Name())

	require.NoError(t, result.Components[0].Component.Stop(ctx))

	rows, err := client.TokenUsage.Query().
		Order(ent.Asc(tokenusage.FieldSessionKey)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "modulesInitDeterministicBranches9-recent", rows[0].SessionKey)
}

func modulesInitDeterministicBranchesComponentNames(entries []lifecycle.ComponentEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Component.Name())
	}
	return names
}

func TestModuleMetadataIncludesAutomationDependencyContracts(t *testing.T) {
	t.Parallel()

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)

	assert.Equal(t, "automation", (&automationModule{cfg: cfg}).Name())
	assert.Equal(t, []appinit.Provides{
		appinit.ProvidesSessionStore,
		appinit.ProvidesRunLedger,
		appinit.ProvidesMission,
	}, (&automationModule{cfg: cfg}).DependsOn())

	assert.Equal(t, []appinit.Provides{
		appinit.ProvidesKnowledge,
		appinit.ProvidesMemory,
		appinit.ProvidesGraph,
		appinit.ProvidesLibrarian,
		appinit.ProvidesSkills,
	}, (&intelligenceModule{cfg: cfg}).Provides())

	assert.Equal(t, []appinit.Provides{
		appinit.ProvidesPayment,
		appinit.ProvidesP2P,
		appinit.ProvidesEconomy,
		appinit.ProvidesContract,
		appinit.ProvidesSmartAccount,
		appinit.ProvidesWorkspace,
	}, (&networkModule{cfg: cfg}).Provides())
}
