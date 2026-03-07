package token

import (
	"time"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/observability"
)

// TokenStore is the interface for persistent token usage storage.
type TokenStore interface {
	Save(usage observability.TokenUsage) error
}

// Tracker subscribes to TokenUsageEvent and forwards data to the
// MetricsCollector and optional persistent store.
type Tracker struct {
	collector *observability.MetricsCollector
	store     TokenStore // nil if persistence disabled
}

// NewTracker creates a new Tracker that records token usage.
func NewTracker(collector *observability.MetricsCollector, store TokenStore) *Tracker {
	return &Tracker{
		collector: collector,
		store:     store,
	}
}

// Subscribe registers the tracker on the event bus.
func (t *Tracker) Subscribe(bus *eventbus.Bus) {
	eventbus.SubscribeTyped[eventbus.TokenUsageEvent](bus, t.handle)
}

func (t *Tracker) handle(evt eventbus.TokenUsageEvent) {
	usage := observability.TokenUsage{
		Provider:     evt.Provider,
		Model:        evt.Model,
		SessionKey:   evt.SessionKey,
		AgentName:    evt.AgentName,
		InputTokens:  evt.InputTokens,
		OutputTokens: evt.OutputTokens,
		TotalTokens:  evt.TotalTokens,
		CacheTokens:  evt.CacheTokens,
		Timestamp:    time.Now(),
	}

	t.collector.RecordTokenUsage(usage)

	if t.store != nil {
		_ = t.store.Save(usage)
	}
}
