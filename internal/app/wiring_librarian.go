package app

import (
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/types"
)

// librarianComponents holds optional proactive librarian components.
type librarianComponents struct {
	inquiryStore    *librarian.InquiryStore
	proactiveBuffer *librarian.ProactiveBuffer
}

// initLibrarian creates the proactive librarian components if enabled.
// Requires: librarian.enabled && knowledge.enabled && observationalMemory.enabled.
func initLibrarian(
	cfg *config.Config,
	sv *supervisor.Supervisor,
	store session.Store,
	kc *knowledgeComponents,
	mc *memoryComponents,
	gc *graphComponents,
	bus *eventbus.Bus,
) (*librarianComponents, *types.FeatureStatus) {
	const featureName = "Librarian"

	if !cfg.Librarian.Enabled {
		logger().Info("proactive librarian disabled")
		return nil, &types.FeatureStatus{Name: featureName, Enabled: false, Healthy: true}
	}
	if kc == nil {
		logger().Warn("proactive librarian requires knowledge system, skipping")
		return nil, &types.FeatureStatus{
			Name: featureName, Enabled: false, Healthy: false,
			Reason:     "requires knowledge system",
			Suggestion: "enable knowledge.enabled to use the librarian",
		}
	}
	if mc == nil {
		logger().Warn("proactive librarian requires observational memory, skipping")
		return nil, &types.FeatureStatus{
			Name: featureName, Enabled: false, Healthy: false,
			Reason:     "requires observational memory",
			Suggestion: "enable observationalMemory.enabled to use the librarian",
		}
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("proactive librarian requires EntStore, skipping")
		return nil, &types.FeatureStatus{
			Name: featureName, Enabled: false, Healthy: false,
			Reason:     "requires EntStore backend",
			Suggestion: "configure session.databasePath with an Ent-backed store",
		}
	}

	client := entStore.Client()
	lLogger := logger()

	inquiryStore := librarian.NewInquiryStore(client, lLogger)

	// Create LLM proxy.
	provider := cfg.Librarian.Provider
	if provider == "" {
		provider = cfg.Agent.Provider
	}
	lModel := cfg.Librarian.Model
	if lModel == "" {
		lModel = cfg.Agent.Model
	}

	proxy := supervisor.NewProviderProxy(sv, provider, lModel)
	generator := &providerTextGenerator{proxy: proxy}

	analyzer := librarian.NewObservationAnalyzer(generator, lLogger)
	processor := librarian.NewInquiryProcessor(generator, inquiryStore, kc.store, lLogger)

	// Message provider.
	getMessages := func(sessionKey string) ([]session.Message, error) {
		sess, err := store.Get(sessionKey)
		if err != nil {
			return nil, err
		}
		if sess == nil {
			return nil, nil
		}
		return sess.History, nil
	}

	// Observation provider.
	getObservations := librarian.ObservationProvider(mc.store.ListObservations)

	bufCfg := librarian.ProactiveBufferConfig{
		ObservationThreshold: cfg.Librarian.ObservationThreshold,
		CooldownTurns:        cfg.Librarian.InquiryCooldownTurns,
		MaxPending:           cfg.Librarian.MaxPendingInquiries,
		AutoSaveConfidence:   cfg.Librarian.AutoSaveConfidence,
	}
	buffer := librarian.NewProactiveBuffer(
		analyzer, processor, inquiryStore, kc.store,
		getMessages, getObservations, bufCfg, lLogger,
	)

	// Wire event bus for graph triple publishing.
	if bus != nil {
		buffer.SetEventBus(bus)
	}

	logger().Infow("proactive librarian initialized",
		"provider", provider,
		"model", lModel,
		"observationThreshold", bufCfg.ObservationThreshold,
		"cooldownTurns", bufCfg.CooldownTurns,
		"maxPending", bufCfg.MaxPending,
	)

	return &librarianComponents{
		inquiryStore:    inquiryStore,
		proactiveBuffer: buffer,
	}, &types.FeatureStatus{Name: featureName, Enabled: true, Healthy: true}
}
