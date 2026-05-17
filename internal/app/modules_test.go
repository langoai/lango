package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ctxkeys"
	entmissionstatehistory "github.com/langoai/lango/internal/ent/missionstatehistory"
	"github.com/langoai/lango/internal/eventbus"
	internalextension "github.com/langoai/lango/internal/extension"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/proposal"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
	"github.com/langoai/lango/internal/toolchain"
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

func TestExtensionModuleInit_SurfaceInvalidMCPProjectConfig(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	t.Setenv("HOME", filepath.Join(tmp, "home"))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, ".lango-mcp.json"), []byte(`{bad json`), 0644))

	cfg := config.DefaultConfig()
	cfg.MCP.Enabled = true

	mod := &extensionModule{cfg: cfg}
	_, err = mod.Init(context.Background(), nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "project")
	assert.Contains(t, err.Error(), ".lango-mcp.json")
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

type missionTestApprovalProvider struct {
	response approval.ApprovalResponse
	err      error
	received *approval.ApprovalRequest
}

func (p *missionTestApprovalProvider) RequestApproval(_ context.Context, req approval.ApprovalRequest) (approval.ApprovalResponse, error) {
	p.received = &req
	if p.err != nil {
		return approval.ApprovalResponse{}, p.err
	}
	return p.response, nil
}

func (p *missionTestApprovalProvider) CanHandle(_ string) bool { return true }

type failingMissionBackgroundLinker struct {
	calls int
}

func (l *failingMissionBackgroundLinker) LinkBackgroundTask(_ context.Context, _ string, _ background.Origin, _ string) error {
	l.calls++
	return assert.AnError
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

func TestIntelligenceModuleInit_ExtSkillCollisionIsFatal(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	extensionsDir := filepath.Join(tmp, "extensions")
	skillsDir := filepath.Join(tmp, "skills")
	writeInstalledExtensionPack(t, extensionsDir, "pack-a", "foo")
	writeInstalledExtensionPack(t, extensionsDir, "pack-b", "foo")
	writeExtensionSkill(t, filepath.Join(skillsDir, "ext-pack-a"), "foo")
	writeExtensionSkill(t, filepath.Join(skillsDir, "ext-pack-b"), "foo")

	extReg, err := internalextension.LoadRegistry(extensionsDir, false)
	require.NoError(t, err)
	require.Len(t, extReg.OKPacks(), 2)

	cfg := config.DefaultConfig()
	cfg.Skill.SkillsDir = skillsDir
	mod := &intelligenceModule{cfg: cfg, extReg: extReg}

	_, err = mod.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "foo")
	assert.Contains(t, err.Error(), "pack-a")
	assert.Contains(t, err.Error(), "pack-b")
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

func writeInstalledExtensionPack(t *testing.T, extensionsDir, packName, skillName string) {
	t.Helper()

	packDir := filepath.Join(extensionsDir, packName)
	skillRel := filepath.ToSlash(filepath.Join("skills", skillName, "SKILL.md"))
	manifest := "schema: lango.extension/v1\n" +
		"name: " + packName + "\n" +
		"version: 0.1.0\n" +
		"description: Test extension pack\n" +
		"contents:\n" +
		"  skills:\n" +
		"    - name: " + skillName + "\n" +
		"      path: " + skillRel + "\n"
	skillMD := []byte("---\nname: " + skillName + "\ndescription: Test skill\nstatus: active\n---\nBody.\n")
	require.NoError(t, os.MkdirAll(filepath.Join(packDir, "skills", skillName), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(packDir, "extension.yaml"), []byte(manifest), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(packDir, filepath.FromSlash(skillRel)), skillMD, 0o644))

	manifestSum := sha256.Sum256([]byte(manifest))
	skillSum := sha256.Sum256(skillMD)
	meta := internalextension.InstalledMeta{
		Name:           packName,
		Version:        "0.1.0",
		ManifestSHA256: hex.EncodeToString(manifestSum[:]),
		FileHashes: map[string]string{
			skillRel: hex.EncodeToString(skillSum[:]),
		},
	}
	data, err := json.Marshal(meta)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(packDir, ".installed"), data, 0o644))
}

func writeExtensionSkill(t *testing.T, root, name string) {
	t.Helper()

	dir := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Test skill\nstatus: active\n---\nBody.\n"), 0o644))
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
		{
			name:     "proposal",
			module:   &proposalModule{},
			wantKeys: []appinit.Provides{appinit.ProvidesProposal},
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

func TestProposalModule_InitProvidesRegistryPreparerAndService(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	mod := &proposalModule{bus: bus}

	require.True(t, mod.Enabled())

	result, err := mod.Init(context.Background(), nil)
	require.NoError(t, err)

	vals, ok := result.Values[appinit.ProvidesProposal].(*proposalValues)
	require.True(t, ok)
	require.NotNil(t, vals)
	require.NotNil(t, vals.registry)
	require.NotNil(t, vals.preparer)
	require.NotNil(t, vals.service)
}

func TestProposalModule_LearningSuggestionEventCreatesAndUpdatesProposal(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	mod := &proposalModule{bus: bus}
	result, err := mod.Init(context.Background(), nil)
	require.NoError(t, err)

	vals := result.Values[appinit.ProvidesProposal].(*proposalValues)
	require.NotNil(t, vals.registry)
	now := time.Now().UTC()

	bus.Publish(eventbus.LearningSuggestionEvent{
		SessionKey:   "sess-1",
		SuggestionID: "suggestion-1",
		Pattern:      "timeout retries",
		ProposedRule: "Use bounded retry",
		Confidence:   0.61,
		Rationale:    "Repeated timeout failures benefited from bounded retry.",
		Timestamp:    now,
	})

	items := vals.registry.ListBySession("sess-1")
	require.Len(t, items, 1)
	assert.Equal(t, proposal.ProposalStatusPrepared, items[0].Status)
	require.NotNil(t, items[0].PreparedBrief)
	assert.Equal(t, "Apply learning rule: Use bounded retry", items[0].Title)

	bus.Publish(eventbus.LearningSuggestionEvent{
		SessionKey:   "sess-1",
		SuggestionID: "suggestion-1",
		Pattern:      "timeout retries",
		ProposedRule: "Use bounded retry",
		Confidence:   0.84,
		Rationale:    "Updated rationale.",
		Timestamp:    now.Add(5 * time.Minute),
	})

	items = vals.registry.ListBySession("sess-1")
	require.Len(t, items, 1)
	assert.Equal(t, 0.84, items[0].Confidence)
	assert.Equal(t, "Updated rationale.", items[0].Reason)
}

func TestProposalModule_DeferredProducersStayInactive(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	mod := &proposalModule{bus: bus}
	result, err := mod.Init(context.Background(), nil)
	require.NoError(t, err)

	vals := result.Values[appinit.ProvidesProposal].(*proposalValues)

	bus.Publish(eventbus.SpecDriftDetectedEvent{
		ToolName:     "exec",
		ErrorClass:   "timeout",
		Occurrences:  3,
		SampleError:  "timed out",
		AffectedSpec: "spec-a",
		Timestamp:    time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
	})

	assert.Empty(t, vals.registry.ListBySession("sess-1"))
	assert.Empty(t, vals.registry.ListBySession("librarian-session"))
}

func TestMissionApprovalObserver_GrantedTransitionsMissionBackToActive(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	store := mission.NewEntStore(client)
	service := mission.NewService(store)
	hooks := &missionApprovalHooks{service: service}
	ctx := context.Background()

	row, err := service.StartMission(ctx, mission.StartMissionInput{
		SessionKey:  "sess-approval-granted",
		Title:       "Approve mission-bound filesystem work",
		StartActive: true,
	})
	require.NoError(t, err)
	require.NoError(t, service.AttachExecution(ctx, mission.AttachExecutionInput{
		MissionID:     row.ID.String(),
		ExecutionKind: mission.ExecutionKindTaskOSExecution,
		ExecutionRef:  "task-approval-1",
		LinkRole:      mission.LinkRolePrimary,
	}))

	provider := &missionTestApprovalProvider{
		response: approval.ApprovalResponse{Approved: true, Provider: "tui"},
	}
	mw := toolchain.WithApproval(
		config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyAll},
		provider,
		nil,
		nil,
		nil,
		hooks,
	)
	tool := &agent.Tool{
		Name:        "exec",
		SafetyLevel: agent.SafetyLevelDangerous,
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	wrapped := toolchain.Chain(tool, mw)

	runCtx := session.WithRunContext(
		session.WithSessionKey(ctxkeys.WithMissionID(context.Background(), row.ID.String()), "sess-approval-granted"),
		session.RunContext{SessionType: "background", RunID: "task-approval-1"},
	)
	result, err := wrapped.Handler(runCtx, map[string]interface{}{"command": "pwd"})
	require.NoError(t, err)
	assert.Equal(t, "ok", result)
	require.NotNil(t, provider.received)
	assert.Equal(t, row.ID.String(), provider.received.MissionID)
	assert.Equal(t, "task_os_execution", provider.received.ExecutionKind)
	assert.Equal(t, "task-approval-1", provider.received.ExecutionRef)

	latest, err := store.GetMission(ctx, row.ID.String())
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, mission.StatusActive, latest.Status)
	assert.Nil(t, latest.CurrentDecisionKind)
	assert.Nil(t, latest.CurrentDecisionSummary)
}

func TestMissionApprovalObserver_DeniedAndTimeoutRemainWaitingDecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		response       approval.ApprovalResponse
		err            error
		wantSummarySub string
	}{
		{
			name:           "denied",
			response:       approval.ApprovalResponse{Approved: false, Provider: "tui"},
			wantSummarySub: "denied",
		},
		{
			name:           "timed out",
			err:            approval.WrapError(approval.ErrTimeout, "tui", "req-timeout", "approval timeout"),
			wantSummarySub: "timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := testutil.TestEntClient(t)
			store := mission.NewEntStore(client)
			service := mission.NewService(store)
			hooks := &missionApprovalHooks{service: service}
			ctx := context.Background()

			row, err := service.StartMission(ctx, mission.StartMissionInput{
				SessionKey:  "sess-approval-" + tt.name,
				Title:       "Mission waiting on operator decision",
				StartActive: true,
			})
			require.NoError(t, err)
			require.NoError(t, service.AttachExecution(ctx, mission.AttachExecutionInput{
				MissionID:     row.ID.String(),
				ExecutionKind: mission.ExecutionKindTaskOSExecution,
				ExecutionRef:  "task-" + tt.name,
				LinkRole:      mission.LinkRolePrimary,
			}))

			provider := &missionTestApprovalProvider{response: tt.response, err: tt.err}
			mw := toolchain.WithApproval(
				config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyAll},
				provider,
				nil,
				nil,
				nil,
				hooks,
			)
			tool := &agent.Tool{
				Name:        "exec",
				SafetyLevel: agent.SafetyLevelDangerous,
				Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
					return "ok", nil
				},
			}
			wrapped := toolchain.Chain(tool, mw)

			runCtx := session.WithRunContext(
				session.WithSessionKey(ctxkeys.WithMissionID(context.Background(), row.ID.String()), "sess-approval-"+tt.name),
				session.RunContext{SessionType: "background", RunID: "task-" + tt.name},
			)
			_, err = wrapped.Handler(runCtx, map[string]interface{}{"command": "pwd"})
			require.Error(t, err)
			require.NotNil(t, provider.received)
			assert.Equal(t, row.ID.String(), provider.received.MissionID)
			assert.Equal(t, "task_os_execution", provider.received.ExecutionKind)
			assert.Equal(t, "task-"+tt.name, provider.received.ExecutionRef)

			latest, err := store.GetMission(ctx, row.ID.String())
			require.NoError(t, err)
			require.NotNil(t, latest)
			assert.Equal(t, mission.StatusWaitingDecision, latest.Status)
			require.NotNil(t, latest.CurrentDecisionSummary)
			assert.Contains(t, *latest.CurrentDecisionSummary, tt.wantSummarySub)

			historyRows, err := client.MissionStateHistory.Query().
				Where(entmissionstatehistory.MissionID(row.ID)).
				Order(entmissionstatehistory.BySeq()).
				All(ctx)
			require.NoError(t, err)
			require.Len(t, historyRows, 3)
			require.NotNil(t, historyRows[1].FromStatus)
			assert.Equal(t, entmissionstatehistory.FromStatusActive, *historyRows[1].FromStatus)
			assert.Equal(t, entmissionstatehistory.ToStatusWaitingDecision, historyRows[1].ToStatus)
			require.NotNil(t, historyRows[2].FromStatus)
			assert.Equal(t, entmissionstatehistory.FromStatusWaitingDecision, *historyRows[2].FromStatus)
			assert.Equal(t, entmissionstatehistory.ToStatusWaitingDecision, historyRows[2].ToStatus)
		})
	}
}

func TestAutomationModule_MissionExecutionLinkAdapterWiredToBackgroundTools(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Background.Enabled = true
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.WriteThrough = true

	client := testutil.TestEntClient(t)
	store := mission.NewEntStore(client)
	service := mission.NewService(store)
	bgLinker := &missionBackgroundLinkHooks{service: service}
	rlStore := runledger.NewMemoryStore()
	rlVals := &runLedgerValues{
		store: rlStore,
		pev:   runledger.NewPEVEngine(rlStore, runledger.DefaultValidators()),
	}
	missionVals := &missionValues{
		store:            store,
		service:          service,
		backgroundLinker: bgLinker,
	}
	missionRow, err := service.StartMission(context.Background(), mission.StartMissionInput{
		SessionKey: "sess-bg-link",
		Title:      "Background mission link",
	})
	require.NoError(t, err)

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
		resp, err := tool.Handler(ctxkeys.WithMissionID(context.Background(), missionRow.ID.String()), map[string]interface{}{"prompt": "mission task"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		taskID := resp.(map[string]interface{})["task_id"].(string)
		link, linkErr := store.FindExecutionLinkByExecution(context.Background(), mission.ExecutionKindTaskOSExecution, taskID)
		require.NoError(t, linkErr)
		require.NotNil(t, link)
		assert.Equal(t, missionRow.ID, link.MissionID)
		break
	}
	require.True(t, submitFound)
	require.Len(t, bgLinker.taskIDs, 1)
	require.Equal(t, "mission task", bgLinker.prompts[0])
}

func TestRunLedgerModule_MissionExecutionLinkAdapterWiredToToolBuilder(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = true

	client := testutil.TestEntClient(t)
	store := mission.NewEntStore(client)
	service := mission.NewService(store)
	runLinker := &missionRunLedgerLinkHooks{service: service}
	missionVals := &missionValues{
		store:           store,
		service:         service,
		runLedgerLinker: runLinker,
	}
	missionRow, err := service.StartMission(context.Background(), mission.StartMissionInput{
		SessionKey: "sess-run-link",
		Title:      "Run mission link",
	})
	require.NoError(t, err)

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
		resp, err := tool.Handler(ctxkeys.WithMissionID(ctxkeys.WithAgentName(context.Background(), "orchestrator"), missionRow.ID.String()), map[string]interface{}{
			"plan_json":        planJSON,
			"session_key":      "sess-1",
			"original_request": "wire mission",
			"valid_agents":     []string{"operator"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		runID := resp.(map[string]interface{})["run_id"].(string)
		link, linkErr := store.FindExecutionLinkByExecution(context.Background(), mission.ExecutionKindRunLedgerRun, runID)
		require.NoError(t, linkErr)
		require.NotNil(t, link)
		assert.Equal(t, missionRow.ID, link.MissionID)
		break
	}
	require.True(t, createFound)
	require.Len(t, runLinker.runIDs, 1)
	require.Equal(t, "sess-1", runLinker.sessionKeys[0])
}

func TestAutomationModule_AgentSpawnMissionBindingAttachesExecutionLink(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Background.Enabled = true

	client := testutil.TestEntClient(t)
	store := mission.NewEntStore(client)
	service := mission.NewService(store)
	bgLinker := &missionBackgroundLinkHooks{service: service}
	rlStore := runledger.NewMemoryStore()
	rlVals := &runLedgerValues{
		store: rlStore,
		pev:   runledger.NewPEVEngine(rlStore, runledger.DefaultValidators()),
	}
	missionVals := &missionValues{
		store:            store,
		service:          service,
		backgroundLinker: bgLinker,
	}
	missionRow, err := service.StartMission(context.Background(), mission.StartMissionInput{
		SessionKey: "sess-agent-link",
		Title:      "Spawn child under mission",
	})
	require.NoError(t, err)

	mod := &automationModule{cfg: cfg, app: &App{Config: cfg}}
	result, err := mod.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{},
		appinit.ProvidesMission:    missionVals,
		appinit.ProvidesRunLedger:  rlVals,
	})
	require.NoError(t, err)

	var spawnFound bool
	for _, tool := range result.Tools {
		if tool.Name != "agent_spawn" {
			continue
		}
		spawnFound = true
		resp, err := tool.Handler(
			ctxkeys.WithMissionID(session.WithSessionKey(context.Background(), "sess-agent-link"), missionRow.ID.String()),
			map[string]interface{}{
				"instruction": "review the logs",
				"agent":       "operator",
			},
		)
		require.NoError(t, err)
		payload := resp.(map[string]interface{})
		agentID := payload["agent_id"].(string)
		link, linkErr := store.FindExecutionLinkByExecution(context.Background(), mission.ExecutionKindTaskOSExecution, agentID)
		require.NoError(t, linkErr)
		require.NotNil(t, link)
		assert.Equal(t, missionRow.ID, link.MissionID)
		break
	}

	require.True(t, spawnFound)
	require.Len(t, bgLinker.taskIDs, 1)
}

func TestAutomationModule_AgentSpawnMissionBindingLinkFailureCancelsSubmittedWork(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Background.Enabled = true

	failingLinker := &failingMissionBackgroundLinker{}
	rlStore := runledger.NewMemoryStore()
	rlVals := &runLedgerValues{
		store: rlStore,
		pev:   runledger.NewPEVEngine(rlStore, runledger.DefaultValidators()),
	}
	missionVals := &missionValues{
		backgroundLinker: failingLinker,
	}

	mod := &automationModule{cfg: cfg, app: &App{Config: cfg}}
	result, err := mod.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{},
		appinit.ProvidesMission:    missionVals,
		appinit.ProvidesRunLedger:  rlVals,
	})
	require.NoError(t, err)

	var (
		spawnFound     bool
		automationVals *automationValues
	)
	automationVals, _ = result.Values[appinit.ProvidesAutomation].(*automationValues)
	require.NotNil(t, automationVals)

	for _, tool := range result.Tools {
		if tool.Name != "agent_spawn" {
			continue
		}
		spawnFound = true
		resp, err := tool.Handler(
			ctxkeys.WithMissionID(session.WithSessionKey(context.Background(), "sess-agent-fail"), "mission-agent-fail"),
			map[string]interface{}{
				"instruction": "review the logs",
				"agent":       "operator",
			},
		)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "attach spawned child execution to mission")
		break
	}

	require.True(t, spawnFound)
	require.Equal(t, 1, failingLinker.calls)
	bgMgr, ok := automationVals.BackgroundManager.(*background.Manager)
	require.True(t, ok)
	snapshots := bgMgr.List()
	require.Len(t, snapshots, 1)
	assert.Equal(t, background.Cancelled, snapshots[0].Status)
}
