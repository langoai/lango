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

func TestPopulateAppFieldsMapsNetworkExtensionObservabilityAndProvenance(t *testing.T) {
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

func TestAppStopTreatsResourceCloseErrorsAsBestEffortCleanup(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("close failed")
	tracerErr := errors.New("tracer shutdown failed")
	ctx, cancel := context.WithCancel(context.Background())
	application := &App{
		ctx:            ctx,
		cancel:         cancel,
		registry:       lifecycle.NewRegistry(),
		Browser:        &populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorCloser{err: closeErr},
		Store:          &populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorSessionStore{err: closeErr},
		GraphStore:     &populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore{err: closeErr},
		TracerShutdown: func(context.Context) error { return tracerErr },
	}

	err := application.Stop(context.Background())

	require.NoError(t, err)
	assert.ErrorIs(t, application.ctx.Err(), context.Canceled)
	assert.True(t, application.Browser.(*populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorCloser).closed)
	assert.True(t, application.Store.(*populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorSessionStore).closed)
	assert.True(t, application.GraphStore.(*populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore).closed)
}

type populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorCloser struct {
	err    error
	closed bool
}

func (c *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorCloser) Close() error {
	c.closed = true
	return c.err
}

type populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorSessionStore struct {
	stubSessionStore
	err    error
	closed bool
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorSessionStore) Close() error {
	s.closed = true
	return s.err
}

type populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore struct {
	err    error
	closed bool
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) AddTriple(context.Context, graph.Triple) error {
	return nil
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) AddTriples(context.Context, []graph.Triple) error {
	return nil
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) RemoveTriple(context.Context, graph.Triple) error {
	return nil
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) QueryBySubject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) QueryByObject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) QueryBySubjectPredicate(
	context.Context,
	string,
	string,
) ([]graph.Triple, error) {
	return nil, nil
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) Traverse(
	context.Context,
	string,
	int,
	[]string,
) ([]graph.Triple, error) {
	return nil, nil
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) Count(context.Context) (int, error) {
	return 0, nil
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) PredicateStats(context.Context) (map[string]int, error) {
	return nil, nil
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) AllTriples(context.Context) ([]graph.Triple, error) {
	return nil, nil
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) ClearAll(context.Context) error {
	return nil
}

func (s *populateAppFieldsMapsNetworkExtensionObservabilityAndProvenanceErrorGraphStore) Close() error {
	s.closed = true
	return s.err
}
