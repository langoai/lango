package app

import (
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/types"
)

// memoryComponents holds optional observational memory components.
type memoryComponents struct {
	store     *memory.Store
	observer  *memory.Observer
	reflector *memory.Reflector
	buffer    *memory.Buffer
}

// initMemory creates the observational memory components if enabled.
func initMemory(cfg *config.Config, store session.Store, sv *supervisor.Supervisor, bus *eventbus.Bus) (*memoryComponents, *types.FeatureStatus) {
	const featureName = "Obs. Memory"

	if !cfg.ObservationalMemory.Enabled {
		logger().Info("observational memory disabled")
		return nil, &types.FeatureStatus{Name: featureName, Enabled: false, Healthy: true}
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("observational memory requires EntStore, skipping")
		return nil, &types.FeatureStatus{
			Name: featureName, Enabled: false, Healthy: false,
			Reason:     "requires EntStore backend",
			Suggestion: "configure session.databasePath with an Ent-backed store",
		}
	}

	client := entStore.Client()
	mLogger := logger()
	mStore := memory.NewStore(client, mLogger)
	if bus != nil {
		mStore.SetEventBus(bus)
	}

	// Create provider proxy for observer/reflector LLM calls
	provider := cfg.ObservationalMemory.Provider
	if provider == "" {
		provider = cfg.Agent.Provider
	}
	omModel := cfg.ObservationalMemory.Model
	if omModel == "" {
		omModel = cfg.Agent.Model
	}

	proxy := supervisor.NewProviderProxy(sv, provider, omModel)
	generator := &providerTextGenerator{proxy: proxy}

	observer := memory.NewObserver(generator, mStore, mLogger)
	reflector := memory.NewReflector(generator, mStore, mLogger)

	// Apply defaults for thresholds
	msgThreshold := cfg.ObservationalMemory.MessageTokenThreshold
	if msgThreshold <= 0 {
		msgThreshold = 1000
	}
	obsThreshold := cfg.ObservationalMemory.ObservationTokenThreshold
	if obsThreshold <= 0 {
		obsThreshold = 2000
	}

	// Message provider retrieves messages for a session key
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

	buffer := memory.NewBuffer(observer, reflector, mStore, msgThreshold, obsThreshold, getMessages, mLogger)

	if cfg.ObservationalMemory.ReflectionConsolidationThreshold > 0 {
		buffer.SetReflectionConsolidationThreshold(cfg.ObservationalMemory.ReflectionConsolidationThreshold)
	}

	logger().Infow("observational memory initialized",
		"provider", provider,
		"model", omModel,
		"messageTokenThreshold", msgThreshold,
		"observationTokenThreshold", obsThreshold,
	)

	return &memoryComponents{
		store:     mStore,
		observer:  observer,
		reflector: reflector,
		buffer:    buffer,
	}, &types.FeatureStatus{Name: featureName, Enabled: true, Healthy: true}
}
