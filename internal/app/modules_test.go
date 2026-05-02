package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModuleTopoSort_AllDisabled verifies that when all optional modules are disabled,
// the build succeeds with only the foundation module.
func TestModuleTopoSort_AllDisabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	modules := []appinit.Module{
		&foundationModule{cfg: cfg},
		&intelligenceModule{cfg: cfg},
		&automationModule{cfg: cfg},
		&networkModule{cfg: cfg},
		&extensionModule{cfg: cfg},
	}

	sorted, err := appinit.TopoSort(modules)
	require.NoError(t, err)
	require.NotEmpty(t, sorted)

	// Foundation should come first (no dependencies).
	assert.Equal(t, "foundation", sorted[0].Name())
}

// TestModuleTopoSort_DependencyOrder verifies that the intelligence module
// comes after the foundation module.
func TestModuleTopoSort_DependencyOrder(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Knowledge.Enabled = true

	modules := []appinit.Module{
		&intelligenceModule{cfg: cfg},
		&foundationModule{cfg: cfg},
		&automationModule{cfg: cfg},
		&extensionModule{cfg: cfg},
	}

	sorted, err := appinit.TopoSort(modules)
	require.NoError(t, err)

	names := make([]string, len(sorted))
	for i, m := range sorted {
		names[i] = m.Name()
	}

	// Foundation must come before intelligence.
	foundIdx := indexOf(names, "foundation")
	intelIdx := indexOf(names, "intelligence")
	assert.True(t, foundIdx < intelIdx, "foundation should come before intelligence: %v", names)
}

// TestModuleEnabled_Automation verifies that the automation module is disabled
// when all automation subsystems are disabled.
func TestModuleEnabled_Automation(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	// All disabled by default.
	m := &automationModule{cfg: cfg}
	assert.False(t, m.Enabled())

	cfg2 := config.DefaultConfig()
	cfg2.Cron.Enabled = true
	m2 := &automationModule{cfg: cfg2}
	assert.True(t, m2.Enabled())
}

// TestModuleBuild_FoundationOnly verifies that foundation module initializes
// successfully when other modules are disabled.
func TestModuleBuild_FoundationOnly(t *testing.T) {
	// This test would require a bootstrap.Result which needs DB setup.
	// Skipping for unit tests — validated in integration tests.
	t.Skip("requires bootstrap.Result with DB client")
}

// TestFoundationCatalogEntries verifies catalog entry generation.
func TestFoundationCatalogEntries(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	entries := buildFoundationCatalogEntries(cfg, nil, nil, nil)

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Category] = true
	}

	assert.True(t, names["exec"])
	assert.True(t, names["filesystem"])
	assert.True(t, names["browser"])
	assert.True(t, names["crypto"])
	assert.True(t, names["secrets"])
}

func indexOf(s []string, target string) int {
	for i, v := range s {
		if v == target {
			return i
		}
	}
	return -1
}

// TestModuleBuild_DisabledModuleDependency verifies that disabled modules
// don't block the initialization of modules that depend on them.
func TestModuleBuild_DisabledModuleDependency(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	// Network depends on Security/SessionStore.
	// When network is disabled, the build should still succeed.
	modules := []appinit.Module{
		&foundationModule{cfg: cfg},
		&networkModule{cfg: cfg}, // disabled (payment/p2p/economy all false)
	}

	sorted, err := appinit.TopoSort(modules)
	require.NoError(t, err)
	// Only foundation should be in sorted (network is disabled).
	require.Len(t, sorted, 1)
	assert.Equal(t, "foundation", sorted[0].Name())
}

// TestExtensionModule_AlwaysEnabled verifies that the extension module is
// always enabled.
func TestExtensionModule_AlwaysEnabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	m := &extensionModule{cfg: cfg}
	assert.True(t, m.Enabled())
}

// TestIntelligenceModule_AlwaysEnabled verifies that the intelligence module is
// always enabled (individual subsystems check their own config).
func TestIntelligenceModule_AlwaysEnabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	m := &intelligenceModule{cfg: cfg}
	assert.True(t, m.Enabled())
}

func TestIntelligenceModule_BuildRegistersReceiptsToolWhenKnowledgeEnabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Knowledge.Enabled = true
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")
	cfg.Agent.Provider = ""
	cfg.Providers = map[string]config.ProviderConfig{
		"google": {
			Type:   "gemini",
			APIKey: "test-key",
		},
	}

	boot := testBoot(t, cfg)
	builder := appinit.NewBuilder().
		AddModule(&foundationModule{cfg: cfg, boot: boot}).
		AddModule(&intelligenceModule{cfg: cfg, boot: boot})

	result, err := builder.Build(context.Background())
	require.NoError(t, err)

	tool := findTool(result.Tools, "create_dispute_ready_receipt")
	require.NotNil(t, tool)
	assert.Equal(t, "knowledge", tool.Capability.Category)
}

func TestModuleBuild_WithEconomyEscrow_RegistersExecuteEscrowRecommendationTool(t *testing.T) {
	t.Parallel()

	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1"}`))
	}))
	defer rpcServer.Close()

	cfg := config.DefaultConfig()
	cfg.Knowledge.Enabled = true
	cfg.Payment.Enabled = true
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = true
	cfg.Security.Signer.Provider = "local"
	cfg.Payment.Network.RPCURL = rpcServer.URL
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")
	cfg.Agent.Provider = ""
	cfg.Providers = map[string]config.ProviderConfig{
		"google": {
			Type:   "gemini",
			APIKey: "test-key",
		},
	}

	crypto := security.NewLocalCryptoProvider()
	require.NoError(t, crypto.Initialize("password123"))

	client := testutil.TestEntClient(t)
	boot := &bootstrap.Result{
		Config:  cfg,
		Crypto:  crypto,
		Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
	}
	bus := eventbus.New()

	builder := appinit.NewBuilder().
		AddModule(&foundationModule{cfg: cfg, boot: boot}).
		AddModule(&networkModule{cfg: cfg, boot: boot, bus: bus}).
		AddModule(&intelligenceModule{cfg: cfg, boot: boot, bus: bus})

	result, err := builder.Build(context.Background())
	require.NoError(t, err)

	tool := findTool(result.Tools, "execute_escrow_recommendation")
	require.NotNil(t, tool)
	assert.Equal(t, "knowledge", tool.Capability.Category)
}

// TestModuleProvides verifies that each module declares its provides keys correctly.
func TestModuleProvides(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	tests := []struct {
		name     string
		module   appinit.Module
		wantKeys []appinit.Provides
	}{
		{
			name:     "foundation",
			module:   &foundationModule{cfg: cfg},
			wantKeys: []appinit.Provides{appinit.ProvidesSupervisor, appinit.ProvidesSessionStore, appinit.ProvidesSecurity},
		},
		{
			name:   "intelligence",
			module: &intelligenceModule{cfg: cfg},
			wantKeys: []appinit.Provides{
				appinit.ProvidesKnowledge, appinit.ProvidesMemory,
				appinit.ProvidesGraph,
				appinit.ProvidesLibrarian, appinit.ProvidesSkills,
			},
		},
		{
			name:     "automation",
			module:   &automationModule{cfg: cfg},
			wantKeys: []appinit.Provides{appinit.ProvidesAutomation},
		},
		{
			name:     "mission",
			module:   &missionModule{},
			wantKeys: []appinit.Provides{appinit.ProvidesMission},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantKeys, tt.module.Provides())
		})
	}
}

func TestRunLedgerModule_WorkspaceIsolationGate(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)

	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.WorkspaceIsolation = true

	mod := &runLedgerModule{
		cfg: cfg,
		boot: &bootstrap.Result{
			Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
		},
	}

	result, err := mod.Init(context.Background(), nil)
	require.NoError(t, err)

	vals, ok := result.Values[appinit.ProvidesRunLedger].(*runLedgerValues)
	require.True(t, ok)
	require.NotNil(t, vals)
	require.NotNil(t, vals.pev)
	assert.True(t, vals.pev.WorkspaceEnabled())
}

func TestAutomationModule_WrapsAgentRunStoreWithRunLedgerMirrorWhenWriteThroughEnabled(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.WriteThrough = true
	cfg.Background.Enabled = true

	boot := &bootstrap.Result{
		Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
	}

	runLedgerVals := &runLedgerValues{store: boot.Storage.RunLedger()}
	store := newAutomationAgentRunStore(cfg, runLedgerVals, nil)
	_, ok := store.(*agentrt.RunLedgerMirrorStore)
	require.True(t, ok)
}

func TestAutomationModule_InitRetainsAgentRunStoreInAutomationValues(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Background.Enabled = true

	mod := &automationModule{cfg: cfg, app: &App{}}
	result, err := mod.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{},
	})
	require.NoError(t, err)

	vals, ok := result.Values[appinit.ProvidesAutomation].(*automationValues)
	require.True(t, ok)
	require.NotNil(t, vals)
	require.NotNil(t, vals.AgentRunStore)
}

func TestWithEntClientMissionAccessor(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	facade := storage.NewFacade(nil, nil, storage.WithEntClient(client))

	store, ok := storage.ResolveEntBacked(facade, mission.NewEntStore)
	require.True(t, ok)
	require.NotNil(t, store)
}

func TestMissionModule_InitProvidesStoreAndService(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	boot := &bootstrap.Result{
		Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
	}

	mod := &missionModule{boot: boot}
	require.True(t, mod.Enabled())

	result, err := mod.Init(context.Background(), nil)
	require.NoError(t, err)

	vals, ok := result.Values[appinit.ProvidesMission].(*missionValues)
	require.True(t, ok)
	require.NotNil(t, vals)
	require.NotNil(t, vals.store)
	require.NotNil(t, vals.service)
	require.NotNil(t, vals.approvalObserver)
	require.NotNil(t, vals.backgroundLinker)
	require.NotNil(t, vals.runLedgerLinker)
}

func TestMissionModule_DisabledWithoutDurableStorage(t *testing.T) {
	t.Parallel()

	assert.False(t, (&missionModule{}).Enabled())
	assert.False(t, (&missionModule{boot: &bootstrap.Result{}}).Enabled())
	assert.False(t, (&missionModule{boot: &bootstrap.Result{Storage: storage.NewFacade(nil, nil)}}).Enabled())
}

func TestAutomationModule_MissionExecutionLinkAdapterWiredToBackgroundTools(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Background.Enabled = true
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.WriteThrough = true

	client := testutil.TestEntClient(t)
	store := mission.NewEntStore(client)
	bgLinker := &missionBackgroundLinkHooks{service: mission.NewService(store)}
	rlStore := runledger.NewMemoryStore()
	rlVals := &runLedgerValues{
		store: rlStore,
		pev:   runledger.NewPEVEngine(rlStore, runledger.DefaultValidators()),
	}
	missionVals := &missionValues{
		store:            store,
		service:          mission.NewService(store),
		backgroundLinker: bgLinker,
	}

	mod := &automationModule{cfg: cfg, app: &App{Config: cfg}}
	result, err := mod.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{},
		appinit.ProvidesMission:    missionVals,
		appinit.ProvidesRunLedger:  rlVals,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	var submitFound bool
	for _, tool := range result.Tools {
		if tool.Name != "bg_submit" {
			continue
		}
		submitFound = true
		resp, err := tool.Handler(context.Background(), map[string]interface{}{"prompt": "mission task"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		break
	}
	require.True(t, submitFound)
	require.Len(t, bgLinker.taskIDs, 1)
	require.Equal(t, "mission task", bgLinker.prompts[0])
}

func TestRunLedgerModule_MissionExecutionLinkAdapterWiredToToolBuilder(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = true

	client := testutil.TestEntClient(t)
	store := mission.NewEntStore(client)
	runLinker := &missionRunLedgerLinkHooks{service: mission.NewService(store)}
	missionVals := &missionValues{
		store:           store,
		service:         mission.NewService(store),
		runLedgerLinker: runLinker,
	}

	mod := &runLedgerModule{cfg: cfg}
	result, err := mod.Init(context.Background(), staticResolver{
		appinit.ProvidesMission: missionVals,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	var createFound bool
	for _, tool := range result.Tools {
		if tool.Name != "run_create" {
			continue
		}
		createFound = true
		planJSON := `{"goal":"wire mission","acceptance_criteria":[],"steps":[{"id":"s1","goal":"do work","owner_agent":"operator","validator":{"type":"build_pass"}}]}`
		resp, err := tool.Handler(ctxkeys.WithAgentName(context.Background(), "orchestrator"), map[string]interface{}{
			"plan_json":        planJSON,
			"session_key":      "sess-1",
			"original_request": "wire mission",
			"valid_agents":     []string{"operator"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		break
	}
	require.True(t, createFound)
	require.Len(t, runLinker.runIDs, 1)
	require.Equal(t, "sess-1", runLinker.sessionKeys[0])
}
