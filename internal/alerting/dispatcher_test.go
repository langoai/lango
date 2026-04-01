package alerting

import (
	"sync"
	"testing"
	"time"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDispatcher_PolicyBlockRate_BelowThreshold(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	d := NewDispatcher(bus, 5, 5)
	d.Subscribe(bus)

	var received []eventbus.AlertEvent
	var mu sync.Mutex
	eventbus.SubscribeTyped[eventbus.AlertEvent](bus, func(evt eventbus.AlertEvent) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, evt)
	})

	// Publish 5 block decisions — exactly at threshold, should not alert.
	for i := 0; i < 5; i++ {
		bus.Publish(eventbus.PolicyDecisionEvent{
			Verdict:    "block",
			SessionKey: "s1",
		})
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Empty(t, received)
}

func TestDispatcher_PolicyBlockRate_ExceedsThreshold(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	d := NewDispatcher(bus, 3, 5)
	d.Subscribe(bus)

	var received []eventbus.AlertEvent
	var mu sync.Mutex
	eventbus.SubscribeTyped[eventbus.AlertEvent](bus, func(evt eventbus.AlertEvent) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, evt)
	})

	// Publish 4 blocks — exceeds threshold of 3.
	for i := 0; i < 4; i++ {
		bus.Publish(eventbus.PolicyDecisionEvent{
			Verdict:    "block",
			SessionKey: "s1",
		})
	}

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, received, 1)
	assert.Equal(t, "policy_block_rate", received[0].Type)
	assert.Equal(t, "warning", received[0].Severity)
	assert.Equal(t, "s1", received[0].SessionKey)
}

func TestDispatcher_Deduplication(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	d := NewDispatcher(bus, 2, 5)
	d.Subscribe(bus)

	var received []eventbus.AlertEvent
	var mu sync.Mutex
	eventbus.SubscribeTyped[eventbus.AlertEvent](bus, func(evt eventbus.AlertEvent) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, evt)
	})

	// Push 10 block decisions — only 1 alert should fire due to dedup.
	for i := 0; i < 10; i++ {
		bus.Publish(eventbus.PolicyDecisionEvent{
			Verdict:    "block",
			SessionKey: "s1",
		})
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, received, 1, "deduplication should suppress duplicate alerts within the same window")
}

func TestDispatcher_IgnoresNonBlockVerdicts(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	d := NewDispatcher(bus, 1, 5)
	d.Subscribe(bus)

	var received []eventbus.AlertEvent
	var mu sync.Mutex
	eventbus.SubscribeTyped[eventbus.AlertEvent](bus, func(evt eventbus.AlertEvent) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, evt)
	})

	// Publish observe and allow verdicts — should not trigger.
	bus.Publish(eventbus.PolicyDecisionEvent{Verdict: "observe"})
	bus.Publish(eventbus.PolicyDecisionEvent{Verdict: "allow"})
	bus.Publish(eventbus.PolicyDecisionEvent{Verdict: "observe"})

	mu.Lock()
	defer mu.Unlock()
	assert.Empty(t, received)
}

func TestPruneWindow(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		give     []time.Time
		wantLen  int
	}{
		{
			give:    []time.Time{now.Add(-6 * time.Minute), now.Add(-4 * time.Minute), now},
			wantLen: 2,
		},
		{
			give:    []time.Time{now.Add(-10 * time.Minute)},
			wantLen: 0,
		},
		{
			give:    []time.Time{now},
			wantLen: 1,
		},
		{
			give:    nil,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		result := pruneWindow(tt.give, now)
		assert.Len(t, result, tt.wantLen)
	}
}
