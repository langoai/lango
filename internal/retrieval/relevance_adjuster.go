package retrieval

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
)

// RelevanceStore defines the narrow interface for relevance score mutations.
// Satisfied by *knowledge.Store.
type RelevanceStore interface {
	BoostRelevanceScore(ctx context.Context, key string, delta, maxScore float64) error
	DecayAllRelevanceScores(ctx context.Context, delta, minScore float64) (int, error)
	ResetAllRelevanceScores(ctx context.Context) (int, error)
}

// RelevanceAdjusterConfig mirrors config.AutoAdjustConfig to avoid import cycle.
type RelevanceAdjusterConfig struct {
	Mode          string  // "shadow" or "active"
	BoostDelta    float64
	DecayDelta    float64
	DecayInterval int
	MinScore      float64
	MaxScore      float64
	WarmupTurns   int
}

// RelevanceAdjuster subscribes to ContextInjectedEvent and optionally mutates
// relevance_score on knowledge entries. Items injected into context get boosted;
// all items decay periodically to prevent score inflation.
//
// v1 signal scope: only old knowledge path items (ContextInjectedEvent.Items).
// Effect scope: primarily LIKE fallback search + coordinator merge priority.
type RelevanceAdjuster struct {
	store     RelevanceStore
	config    RelevanceAdjusterConfig
	turnCount atomic.Int64 // process-local, resets on restart
	logger    *zap.SugaredLogger
}

// NewRelevanceAdjuster creates a relevance adjuster with the given config.
func NewRelevanceAdjuster(store RelevanceStore, cfg RelevanceAdjusterConfig, logger *zap.SugaredLogger) *RelevanceAdjuster {
	return &RelevanceAdjuster{
		store:  store,
		config: cfg,
		logger: logger,
	}
}

// Subscribe registers the adjuster to receive ContextInjectedEvent from the bus.
func (a *RelevanceAdjuster) Subscribe(bus *eventbus.Bus) {
	eventbus.SubscribeTyped[eventbus.ContextInjectedEvent](bus, a.handleContextInjected)
}

func (a *RelevanceAdjuster) handleContextInjected(evt eventbus.ContextInjectedEvent) {
	turn := a.turnCount.Add(1)

	// Warmup: don't adjust until sufficient turns observed.
	if int(turn) <= a.config.WarmupTurns {
		a.logger.Debugw("relevance adjuster warmup",
			"turn", turn, "warmupTurns", a.config.WarmupTurns)
		return
	}

	// Collect unique user_knowledge keys (turn-level dedup).
	knowledgeKeys := make(map[string]struct{})
	for _, item := range evt.Items {
		if item.Layer == "user_knowledge" && item.Key != "" {
			knowledgeKeys[item.Key] = struct{}{}
		}
	}

	if a.config.Mode != "active" {
		a.logger.Infow("relevance adjuster shadow",
			"turn", turn,
			"keys_to_boost", len(knowledgeKeys),
		)
		return
	}

	// Active mode: decay first, boost second.
	ctx := context.Background()

	// STEP 1: Global decay (if interval reached).
	if a.config.DecayInterval > 0 && int(turn)%a.config.DecayInterval == 0 {
		n, err := a.store.DecayAllRelevanceScores(ctx, a.config.DecayDelta, a.config.MinScore)
		if err != nil {
			a.logger.Warnw("relevance decay error", "error", err)
		} else {
			a.logger.Infow("relevance decay applied", "decayed", n, "delta", a.config.DecayDelta)
		}
	}

	// STEP 2: Boost injected items.
	for key := range knowledgeKeys {
		if err := a.store.BoostRelevanceScore(ctx, key, a.config.BoostDelta, a.config.MaxScore); err != nil {
			a.logger.Warnw("relevance boost error", "key", key, "error", err)
		}
	}

	if len(knowledgeKeys) > 0 {
		a.logger.Infow("relevance boost applied",
			"boosted", len(knowledgeKeys), "delta", a.config.BoostDelta)
	}
}

// SetMode changes the adjuster mode at runtime (for rollback toggle).
func (a *RelevanceAdjuster) SetMode(mode string) {
	a.config.Mode = mode
}

// Mode returns the current adjuster mode.
func (a *RelevanceAdjuster) Mode() string {
	return a.config.Mode
}
