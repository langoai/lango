package app

import (
	"context"
	"sync"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/observability/health"
	"github.com/langoai/lango/internal/observability/token"
	"github.com/langoai/lango/internal/toolchain"
)

// observabilityComponents holds optional observability components.
type observabilityComponents struct {
	collector      *observability.MetricsCollector
	healthRegistry *health.Registry
	tracker        *token.Tracker
	tokenStore     *token.EntTokenStore
}

// initObservability creates observability components if enabled.
func initObservability(cfg *config.Config, dbClient *ent.Client, bus *eventbus.Bus) *observabilityComponents {
	if !cfg.Observability.Enabled {
		logger().Info("observability disabled")
		return nil
	}

	oc := &observabilityComponents{}

	// 1. Metrics Collector (always created when observability is enabled)
	oc.collector = observability.NewCollector()
	logger().Info("observability: metrics collector initialized")

	// 2. Health Registry
	if cfg.Observability.Health.Enabled {
		oc.healthRegistry = health.NewRegistry()

		// Register built-in memory check (warn at 512 MB)
		oc.healthRegistry.Register(health.NewMemoryCheck(512 * 1024 * 1024))

		logger().Info("observability: health registry initialized")
	}

	// 3. Token Store (persistent, optional)
	if cfg.Observability.Tokens.PersistHistory && dbClient != nil {
		oc.tokenStore = token.NewEntTokenStore(dbClient)
		logger().Info("observability: token store (persistent) initialized")
	}

	// 4. Token Tracker — subscribes to TokenUsageEvent
	if cfg.Observability.Tokens.Enabled {
		var store token.TokenStore
		if oc.tokenStore != nil {
			store = oc.tokenStore
		}
		oc.tracker = token.NewTracker(oc.collector, store)
		oc.tracker.Subscribe(bus)
		logger().Info("observability: token tracker subscribed to event bus")
	}

	// 5. Subscribe to ToolExecutedEvent for tool metrics
	eventbus.SubscribeTyped[toolchain.ToolExecutedEvent](bus, func(evt toolchain.ToolExecutedEvent) {
		oc.collector.RecordToolExecution(evt.ToolName, evt.AgentName, evt.Duration, evt.Success)
	})
	logger().Info("observability: tool execution metrics wired")

	return oc
}

// wireModelAdapterTokenUsage sets up the OnTokenUsage callback on the model adapter
// so token usage events are published to the event bus.
func wireModelAdapterTokenUsage(adapter *adk.ModelAdapter, bus *eventbus.Bus) {
	if adapter == nil || bus == nil {
		return
	}
	adapter.OnTokenUsage = func(providerID, model string, input, output, total, cache int64) {
		bus.Publish(eventbus.TokenUsageEvent{
			Provider:     providerID,
			Model:        model,
			InputTokens:  input,
			OutputTokens: output,
			TotalTokens:  total,
			CacheTokens:  cache,
		})
	}
}

// registerObservabilityLifecycle registers observability components with the lifecycle registry.
func registerObservabilityLifecycle(reg *lifecycle.Registry, oc *observabilityComponents, cfg *config.Config) {
	if oc == nil {
		return
	}

	// Token store cleanup on shutdown
	if oc.tokenStore != nil && cfg.Observability.Tokens.RetentionDays > 0 {
		retDays := cfg.Observability.Tokens.RetentionDays
		store := oc.tokenStore
		reg.Register(lifecycle.NewFuncComponent("observability-token-cleanup",
			func(_ context.Context, _ *sync.WaitGroup) error { return nil },
			func(ctx context.Context) error {
				count, err := store.Cleanup(ctx, retDays)
				if err != nil {
					logger().Warnw("token usage cleanup", "error", err)
					return nil
				}
				if count > 0 {
					logger().Infow("token usage cleanup", "deleted", count, "retentionDays", retDays)
				}
				return nil
			},
		), lifecycle.PriorityCore)
	}
}
