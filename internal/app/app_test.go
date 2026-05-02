package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/proposal"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type staticResolver map[appinit.Provides]interface{}

func (r staticResolver) Resolve(key appinit.Provides) interface{} {
	return r[key]
}

type appTestApprovalProvider struct {
	response approval.ApprovalResponse
}

func (p *appTestApprovalProvider) RequestApproval(_ context.Context, _ approval.ApprovalRequest) (approval.ApprovalResponse, error) {
	return p.response, nil
}

func (p *appTestApprovalProvider) CanHandle(_ string) bool { return true }

// testBoot creates a minimal bootstrap.Result for testing.
func testBoot(t *testing.T, cfg *config.Config) *bootstrap.Result {
	t.Helper()
	return &bootstrap.Result{
		Config: cfg,
	}
}

func TestNew_MinimalConfig(t *testing.T) {
	t.Skip("requires provider credentials; run manually with GOOGLE_API_KEY set")

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "google"
	cfg.Agent.Model = "gemini-2.0-flash"
	cfg.Providers = map[string]config.ProviderConfig{
		"google": {
			Type:   "gemini",
			APIKey: "test-key",
		},
	}

	app, err := New(testBoot(t, cfg))
	require.NoError(t, err)
	require.NotNil(t, app.Agent, "expected agent to be initialized")
	require.NotNil(t, app.Gateway, "expected gateway to be initialized")
	require.NotNil(t, app.Store, "expected store to be initialized")
}

func TestNew_SecurityDisabledByDefault(t *testing.T) {
	t.Skip("requires provider credentials; run manually with GOOGLE_API_KEY set")

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "google"
	cfg.Providers = map[string]config.ProviderConfig{
		"google": {
			Type:   "gemini",
			APIKey: "test-key",
		},
	}

	// Security is not configured — should not block startup
	_, err := New(testBoot(t, cfg))
	require.NoError(t, err, "New() should succeed without security config")
}

func TestNew_NoProviders(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = nil
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")
	_, err := New(testBoot(t, cfg))
	require.Error(t, err, "expected error when no providers configured")
}

func TestNew_InvalidProviderType(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"test": {Type: "nonexistent", APIKey: "test-key"},
	}
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")
	_, err := New(testBoot(t, cfg))
	require.Error(t, err, "expected error for invalid provider type")
}

// ── Phase B Cleanup Stack Tests ──

func TestCleanupStack_RollbackReverseOrder(t *testing.T) {
	var order []string
	var s cleanupStack

	s.push("step-A", func() { order = append(order, "A") })
	s.push("step-B", func() { order = append(order, "B") })
	s.push("step-C", func() { order = append(order, "C") })

	s.rollback()

	assert.Equal(t, []string{"C", "B", "A"}, order, "cleanups must execute in reverse order")
	assert.Empty(t, s.entries, "stack should be empty after rollback")
}

func TestCleanupStack_ClearDiscardsWithoutExecution(t *testing.T) {
	executed := false
	var s cleanupStack

	s.push("step-A", func() { executed = true })
	s.clear()

	assert.False(t, executed, "clear must not execute cleanup functions")
	assert.Empty(t, s.entries, "stack should be empty after clear")
}

func TestCleanupStack_RollbackEmpty(t *testing.T) {
	var s cleanupStack
	// Should not panic on empty stack.
	s.rollback()
	assert.Empty(t, s.entries)
}

func TestCleanupStack_PushAndRollbackPartial(t *testing.T) {
	var order []string
	var s cleanupStack

	s.push("output-store", func() { order = append(order, "output-store") })
	s.push("gateway", func() { order = append(order, "gateway") })

	// Simulate B6 failure — rollback should clean up gateway then output-store.
	s.rollback()

	assert.Equal(t, []string{"gateway", "output-store"}, order,
		"B6 failure should rollback gateway then output-store")
}

func TestNew_PhaseBRollback_AgentCreationFailure(t *testing.T) {
	// No providers configured: Phase A succeeds (supervisor with zero providers),
	// but B6 (initAgent) fails — triggering Phase B rollback of OutputStore + Gateway.
	cfg := config.DefaultConfig()
	cfg.Providers = nil
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")

	_, err := New(testBoot(t, cfg))
	require.Error(t, err, "expected error when agent creation fails")
}

func TestPopulateAppFields_RunLedgerStoreStillAvailable(t *testing.T) {
	t.Parallel()

	app := &App{}
	store := runledger.NewMemoryStore()

	populateAppFields(app, staticResolver{
		appinit.ProvidesRunLedger: &runLedgerValues{store: store},
	})

	assert.Same(t, store, app.RunLedgerStore)
}

func TestPopulateAppFields_AgentRunStoreFromAutomation(t *testing.T) {
	t.Parallel()

	app := &App{}
	store := agentrt.NewInMemoryAgentRunStore()

	populateAppFields(app, staticResolver{
		appinit.ProvidesAutomation: &automationValues{AgentRunStore: store},
	})

	assert.Same(t, store, app.AgentRunStore)
}

func TestPopulateAppFields_AutomationAbsentLeavesAgentRunStoreNil(t *testing.T) {
	t.Parallel()

	app := &App{}

	populateAppFields(app, staticResolver{})

	assert.Nil(t, app.AgentRunStore)
}

func TestPopulateAppFields_MissionComponents(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	store := mission.NewEntStore(client)
	service := mission.NewService(store)
	observer := &missionApprovalHooks{service: service}
	bgLinker := &missionBackgroundLinkHooks{service: service}
	runLinker := &missionRunLedgerLinkHooks{service: service}

	app := &App{}
	populateAppFields(app, staticResolver{
		appinit.ProvidesMission: &missionValues{
			store:            store,
			service:          service,
			approvalObserver: observer,
			backgroundLinker: bgLinker,
			runLedgerLinker:  runLinker,
		},
	})

	assert.Same(t, store, app.MissionStore)
	assert.Same(t, service, app.MissionService)
	assert.Same(t, observer, app.missionApprovalObserver)
	assert.Same(t, bgLinker, app.missionBackgroundLinker)
	assert.Same(t, runLinker, app.missionRunLedgerLinker)
}

func TestPopulateAppFields_ProposalComponents(t *testing.T) {
	t.Parallel()

	registry := proposal.NewRegistry(nil)
	preparer := proposal.NewDeterministicPreparer()
	service := proposal.NewService(registry, preparer)

	app := &App{}
	populateAppFields(app, staticResolver{
		appinit.ProvidesProposal: &proposalValues{
			registry: registry,
			preparer: preparer,
			service:  service,
		},
	})

	assert.Same(t, registry, app.ProposalRegistry)
	assert.Same(t, preparer, app.ProposalPreparer)
	assert.Same(t, service, app.ProposalService)
}

func TestNew_MissionApprovalObserverWiredAtCompositionSite(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")
	cfg.Agent.Provider = "google"
	cfg.Providers = map[string]config.ProviderConfig{
		"google": {
			Type:   "gemini",
			APIKey: "test-key",
		},
	}
	cfg.Security.Interceptor.ApprovalPolicy = config.ApprovalPolicyDangerous

	client := testutil.TestEntClient(t)
	boot := &bootstrap.Result{
		Config:  cfg,
		Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
	}

	app, err := New(boot)
	require.NoError(t, err)
	require.NotNil(t, app.MissionService)

	obs, ok := app.missionApprovalObserver.(*missionApprovalHooks)
	require.True(t, ok)
	mw := toolchain.WithApproval(
		cfg.Security.Interceptor,
		&appTestApprovalProvider{response: approval.ApprovalResponse{Approved: true, Provider: "tui"}},
		nil,
		nil,
		nil,
		app.missionApprovalObserver,
	)
	tool := &agent.Tool{
		Name:        "exec",
		SafetyLevel: agent.SafetyLevelDangerous,
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	wrapped := toolchain.Chain(tool, mw)
	_, err = wrapped.Handler(context.Background(), map[string]interface{}{"command": "pwd"})
	require.NoError(t, err)
	require.NotEmpty(t, obs.requests)
	assert.Equal(t, "exec", obs.requests[len(obs.requests)-1])
}

func TestNew_ProposalServiceAvailableForLaterMutationUse(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "test.db")
	cfg.Agent.Provider = "google"
	cfg.Providers = map[string]config.ProviderConfig{
		"google": {
			Type:   "gemini",
			APIKey: "test-key",
		},
	}

	client := testutil.TestEntClient(t)
	boot := &bootstrap.Result{
		Config:  cfg,
		Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
	}

	app, err := New(boot)
	require.NoError(t, err)
	require.NotNil(t, app.ProposalRegistry)
	require.NotNil(t, app.ProposalPreparer)
	require.NotNil(t, app.ProposalService)
}
