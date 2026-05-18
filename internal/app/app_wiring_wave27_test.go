package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/extension"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/mcp"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/observability/health"
	"github.com/langoai/lango/internal/p2p/agentpool"
	"github.com/langoai/lango/internal/provenance"
)

func TestWave27PopulateAppFieldsMapsNetworkExtensionObservabilityAndProvenance(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	wallet := &wiringP2PWallet{}
	pool := agentpool.New(testLog())
	selector := agentpool.NewSelector(pool, agentpool.DefaultWeights())
	provider := agentpool.NewPoolProvider(pool, selector)
	manager := mcp.NewServerManager(config.MCPConfig{})
	collector := observability.NewCollector()
	healthRegistry := health.NewRegistry()
	checkpointStore := provenance.NewMemoryStore()
	checkpointService := provenance.NewCheckpointService(
		checkpointStore,
		nil,
		cfg.Provenance.Checkpoints,
	)
	sessionTree := provenance.NewSessionTree(provenance.NewMemoryTreeStore())
	attribution := provenance.NewAttributionService(
		provenance.NewMemoryAttributionStore(),
		checkpointStore,
		nil,
	)
	extRegistry := &extension.Registry{}

	application := &App{ExtensionRegistry: extRegistry}
	populateAppFields(application, staticResolver{
		appinit.ProvidesPayment: &paymentComponents{wallet: wallet},
		appinit.ProvidesP2P: &p2pComponents{
			agentPool: pool,
			selector:  selector,
			provider:  provider,
		},
		appinit.ProvidesSmartAccount: &smartAccountComponents{},
		appinit.ProvidesMCP:          &mcpComponents{manager: manager},
		appinit.ProvidesObservability: &observabilityComponents{
			collector:      collector,
			healthRegistry: healthRegistry,
			tracerShutdown: func(context.Context) error { return nil },
		},
		appinit.ProvidesProvenance: &provenanceValues{
			checkpointService: checkpointService,
			sessionTree:       sessionTree,
			attribution:       attribution,
		},
	})

	assert.Same(t, wallet, application.WalletProvider)
	assert.Same(t, pool, application.P2PAgentPool)
	assert.Same(t, provider, application.P2PAgentProvider)
	assert.NotNil(t, application.SmartAccountComponents)
	assert.Same(t, manager, application.MCPManager)
	assert.Same(t, collector, application.MetricsCollector)
	assert.Same(t, healthRegistry, application.HealthRegistry)
	assert.NotNil(t, application.TracerShutdown)
	assert.Same(t, checkpointService, application.ProvenanceCheckpoints)
	assert.Same(t, sessionTree, application.ProvenanceSessionTree)
	assert.Same(t, attribution, application.ProvenanceAttribution)
	assert.Same(t, extRegistry, application.ExtensionRegistry)
}

func TestWave27AppStopTreatsResourceCloseErrorsAsBestEffortCleanup(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("close failed")
	tracerErr := errors.New("tracer shutdown failed")
	ctx, cancel := context.WithCancel(context.Background())
	application := &App{
		ctx:            ctx,
		cancel:         cancel,
		registry:       lifecycle.NewRegistry(),
		Browser:        &wave27ErrorCloser{err: closeErr},
		Store:          &wave27ErrorSessionStore{err: closeErr},
		GraphStore:     &wave27ErrorGraphStore{err: closeErr},
		TracerShutdown: func(context.Context) error { return tracerErr },
	}

	err := application.Stop(context.Background())

	require.NoError(t, err)
	assert.ErrorIs(t, application.ctx.Err(), context.Canceled)
	assert.True(t, application.Browser.(*wave27ErrorCloser).closed)
	assert.True(t, application.Store.(*wave27ErrorSessionStore).closed)
	assert.True(t, application.GraphStore.(*wave27ErrorGraphStore).closed)
}

type wave27ErrorCloser struct {
	err    error
	closed bool
}

func (c *wave27ErrorCloser) Close() error {
	c.closed = true
	return c.err
}

type wave27ErrorSessionStore struct {
	stubSessionStore
	err    error
	closed bool
}

func (s *wave27ErrorSessionStore) Close() error {
	s.closed = true
	return s.err
}

type wave27ErrorGraphStore struct {
	err    error
	closed bool
}

func (s *wave27ErrorGraphStore) AddTriple(context.Context, graph.Triple) error {
	return nil
}

func (s *wave27ErrorGraphStore) AddTriples(context.Context, []graph.Triple) error {
	return nil
}

func (s *wave27ErrorGraphStore) RemoveTriple(context.Context, graph.Triple) error {
	return nil
}

func (s *wave27ErrorGraphStore) QueryBySubject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}

func (s *wave27ErrorGraphStore) QueryByObject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}

func (s *wave27ErrorGraphStore) QueryBySubjectPredicate(
	context.Context,
	string,
	string,
) ([]graph.Triple, error) {
	return nil, nil
}

func (s *wave27ErrorGraphStore) Traverse(
	context.Context,
	string,
	int,
	[]string,
) ([]graph.Triple, error) {
	return nil, nil
}

func (s *wave27ErrorGraphStore) Count(context.Context) (int, error) {
	return 0, nil
}

func (s *wave27ErrorGraphStore) PredicateStats(context.Context) (map[string]int, error) {
	return nil, nil
}

func (s *wave27ErrorGraphStore) AllTriples(context.Context) ([]graph.Triple, error) {
	return nil, nil
}

func (s *wave27ErrorGraphStore) ClearAll(context.Context) error {
	return nil
}

func (s *wave27ErrorGraphStore) Close() error {
	s.closed = true
	return s.err
}
