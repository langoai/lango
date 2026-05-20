package app

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildApprovalProvider_HeadlessFallbackDoesNotApproveP2P(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.HeadlessAutoApprove = true
	cfg.P2P.Enabled = true
	gw := initGateway(cfg, nil, &stubSessionStore{}, nil)

	provider, grants := buildApprovalProvider(cfg, gw)
	require.NotNil(t, provider)
	require.NotNil(t, grants)

	resp, err := provider.RequestApproval(context.Background(), approval.ApprovalRequest{
		ID:         "non-p2p",
		ToolName:   "exec_run",
		SessionKey: "local-session",
	})
	require.NoError(t, err)
	assert.True(t, resp.Approved)
	assert.Equal(t, "headless", resp.Provider)

	resp, err = provider.RequestApproval(context.Background(), approval.ApprovalRequest{
		ID:         "p2p",
		ToolName:   "exec_run",
		SessionKey: "p2p:did:peer:remote",
	})
	require.Error(t, err)
	assert.False(t, resp.Approved)
	assert.Contains(t, err.Error(), "TTY approval unavailable")
	assert.False(t, grants.IsGranted("p2p:did:peer:remote", "exec_run"))
}

func TestWirePostAgent_RegistersBackgroundRoutesWithoutOptionalProviders(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	app := &App{
		Config:  cfg,
		Gateway: initGateway(cfg, nil, &stubSessionStore{}, nil),
	}

	wirePostAgent(
		app,
		staticResolver{},
		nil,
		eventbus.New(),
		approval.NewCompositeProvider(),
		approval.NewGrantStore(),
		nil,
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/bg/tasks", nil)
	rec := httptest.NewRecorder()
	app.Gateway.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "background manager unavailable")
}

func TestResolveIntelligenceHelpers_NilAndWrongSkillRegistryAreSafe(t *testing.T) {
	t.Parallel()

	assert.Nil(t, resolveKC(nil))
	assert.Nil(t, resolveMC(nil))
	assert.Nil(t, resolveGC(nil))
	assert.Nil(t, resolveLC(nil))
	assert.Nil(t, resolveSR(nil))
	assert.Nil(t, resolveSR(&intelligenceValues{SkillRegistry: "not-a-registry"}))

	iv := &intelligenceValues{
		KC:            &knowledgeComponents{},
		MC:            &memoryComponents{},
		GC:            &graphComponents{},
		LC:            &librarianComponents{},
		SkillRegistry: nil,
	}
	assert.Same(t, iv.KC, resolveKC(iv))
	assert.Same(t, iv.MC, resolveMC(iv))
	assert.Same(t, iv.GC, resolveGC(iv))
	assert.Same(t, iv.LC, resolveLC(iv))
}

func TestAppStop_CancelsContextClosesResourcesAndReturnsStopError(t *testing.T) {
	t.Parallel()

	stopErr := errors.New("component stop failed")
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		ctx:            ctx,
		cancel:         cancel,
		registry:       lifecycle.NewRegistry(),
		Browser:        &closeTrackingCloser{},
		Store:          &closeTrackingStore{},
		GraphStore:     &closeTrackingGraphStore{},
		TracerShutdown: func(context.Context) error { return nil },
	}
	tracerCalled := false
	app.TracerShutdown = func(context.Context) error {
		tracerCalled = true
		return nil
	}
	app.registry.Register(lifecycle.NewFuncComponent(
		"failing-component",
		func(context.Context, *sync.WaitGroup) error { return nil },
		func(context.Context) error { return stopErr },
	), lifecycle.PriorityCore)
	require.NoError(t, app.Start(context.Background()))

	err := app.Stop(context.Background())

	require.Error(t, err)
	assert.ErrorIs(t, err, stopErr)
	assert.ErrorIs(t, app.ctx.Err(), context.Canceled)
	assert.True(t, app.Browser.(*closeTrackingCloser).closed)
	assert.True(t, app.Store.(*closeTrackingStore).closed)
	assert.True(t, app.GraphStore.(*closeTrackingGraphStore).closed)
	assert.True(t, tracerCalled)
}

func TestAppStop_ReturnsContextErrorWhenWaitGroupDoesNotDrainAfterCancel(t *testing.T) {
	t.Parallel()

	app := &App{
		registry: lifecycle.NewRegistry(),
	}
	app.wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := app.Stop(ctx)
	app.wg.Done()

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestInitSecurity_RPCRequiresEntStore(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = "rpc"

	crypto, keys, secrets, err := initSecurity(cfg, &stubSessionStore{}, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rpc security provider requires EntStore")
	assert.Nil(t, crypto)
	assert.Nil(t, keys)
	assert.Nil(t, secrets)
}

func TestInitAuth_InvalidProviderConfigIsSkipped(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Auth.Providers = map[string]config.OIDCProviderConfig{
		"bad": {
			IssuerURL:    "://not-a-url",
			ClientID:     "client",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
		},
	}

	auth := initAuth(cfg, &stubSessionStore{})

	assert.Nil(t, auth)
}

func TestBuildProvenanceAgentOptions_UpdatesHookMetadataSnapshot(t *testing.T) {
	t.Parallel()

	registry := toolchain.NewHookRegistry()
	registry.RegisterPre(testPreHook{name: "pre", priority: 10})
	registry.RegisterPost(testPostHook{name: "post", priority: 20})
	pv := &provenanceValues{
		sessionTree:    provenance.NewSessionTree(provenance.NewMemoryTreeStore()),
		configMetadata: map[string]string{"config_fingerprint": "abc123"},
	}

	opts := buildProvenanceAgentOptions(pv, registry)

	require.NotEmpty(t, opts)
	assert.Contains(t, pv.configMetadata["hook_registry"], `"name":"pre"`)
	assert.Contains(t, pv.configMetadata["hook_registry"], `"name":"post"`)
}

func TestBuildProvenanceAgentOptions_ReturnsNilWithoutSessionTree(t *testing.T) {
	t.Parallel()

	assert.Nil(t, buildProvenanceAgentOptions(nil, nil))
	assert.Nil(t, buildProvenanceAgentOptions(&provenanceValues{}, nil))
}

type closeTrackingCloser struct {
	closed bool
}

func (c *closeTrackingCloser) Close() error {
	c.closed = true
	return nil
}

var _ io.Closer = (*closeTrackingCloser)(nil)

type closeTrackingStore struct {
	stubSessionStore
	closed bool
}

func (s *closeTrackingStore) Close() error {
	s.closed = true
	return nil
}

type closeTrackingGraphStore struct {
	closed bool
}

func (s *closeTrackingGraphStore) AddTriple(context.Context, graph.Triple) error {
	return nil
}

func (s *closeTrackingGraphStore) AddTriples(context.Context, []graph.Triple) error {
	return nil
}

func (s *closeTrackingGraphStore) RemoveTriple(context.Context, graph.Triple) error {
	return nil
}

func (s *closeTrackingGraphStore) QueryBySubject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}

func (s *closeTrackingGraphStore) QueryByObject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}

func (s *closeTrackingGraphStore) QueryBySubjectPredicate(
	context.Context,
	string,
	string,
) ([]graph.Triple, error) {
	return nil, nil
}

func (s *closeTrackingGraphStore) Traverse(
	context.Context,
	string,
	int,
	[]string,
) ([]graph.Triple, error) {
	return nil, nil
}

func (s *closeTrackingGraphStore) Count(context.Context) (int, error) {
	return 0, nil
}

func (s *closeTrackingGraphStore) PredicateStats(context.Context) (map[string]int, error) {
	return nil, nil
}

func (s *closeTrackingGraphStore) AllTriples(context.Context) ([]graph.Triple, error) {
	return nil, nil
}

func (s *closeTrackingGraphStore) ClearAll(context.Context) error {
	return nil
}

func (s *closeTrackingGraphStore) Close() error {
	s.closed = true
	return nil
}

var _ graph.Store = (*closeTrackingGraphStore)(nil)

type testPreHook struct {
	name     string
	priority int
}

func (h testPreHook) Name() string { return h.name }

func (h testPreHook) Priority() int { return h.priority }

func (h testPreHook) Pre(toolchain.HookContext) (toolchain.PreHookResult, error) {
	return toolchain.PreHookResult{Action: toolchain.Continue}, nil
}

type testPostHook struct {
	name     string
	priority int
}

func (h testPostHook) Name() string { return h.name }

func (h testPostHook) Priority() int { return h.priority }

func (h testPostHook) Post(toolchain.HookContext, interface{}, error) error {
	return nil
}

var _ toolchain.PreToolHook = testPreHook{}
var _ toolchain.PostToolHook = testPostHook{}

var _ session.Store = (*closeTrackingStore)(nil)
