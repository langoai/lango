package retrieval

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
)

type mockRelevanceStore struct {
	boosted    map[string]float64 // key → total delta
	decayed    int                // number of DecayAll calls
	decayDelta float64
	reset      bool
}

func newMockRelevanceStore() *mockRelevanceStore {
	return &mockRelevanceStore{boosted: make(map[string]float64)}
}

func (m *mockRelevanceStore) BoostRelevanceScore(_ context.Context, key string, delta, _ float64) error {
	m.boosted[key] += delta
	return nil
}

func (m *mockRelevanceStore) DecayAllRelevanceScores(_ context.Context, delta, _ float64) (int, error) {
	m.decayed++
	m.decayDelta = delta
	return 5, nil
}

func (m *mockRelevanceStore) ResetAllRelevanceScores(_ context.Context) (int, error) {
	m.reset = true
	return 10, nil
}

func activeConfig() RelevanceAdjusterConfig {
	return RelevanceAdjusterConfig{
		Mode:          "active",
		BoostDelta:    0.05,
		DecayDelta:    0.01,
		DecayInterval: 10,
		MinScore:      0.1,
		MaxScore:      5.0,
		WarmupTurns:   3,
	}
}

func makeEvent(keys ...string) eventbus.ContextInjectedEvent {
	items := make([]eventbus.ContextInjectedItem, 0, len(keys))
	for _, k := range keys {
		items = append(items, eventbus.ContextInjectedItem{
			Layer: "user_knowledge",
			Key:   k,
			Score: 1.0,
		})
	}
	return eventbus.ContextInjectedEvent{Items: items}
}

func TestRelevanceAdjuster_Shadow_NoWrite(t *testing.T) {
	store := newMockRelevanceStore()
	cfg := activeConfig()
	cfg.Mode = "shadow"
	cfg.WarmupTurns = 0

	adj := NewRelevanceAdjuster(store, cfg, zap.NewNop().Sugar())
	adj.handleContextInjected(makeEvent("key1", "key2"))

	assert.Empty(t, store.boosted, "shadow mode should not write")
	assert.Equal(t, 0, store.decayed, "shadow mode should not decay")
}

func TestRelevanceAdjuster_Warmup(t *testing.T) {
	store := newMockRelevanceStore()
	cfg := activeConfig()
	cfg.WarmupTurns = 3

	adj := NewRelevanceAdjuster(store, cfg, zap.NewNop().Sugar())

	// Turns 1-3: warmup, no writes
	for i := 0; i < 3; i++ {
		adj.handleContextInjected(makeEvent("key1"))
	}
	assert.Empty(t, store.boosted, "warmup turns should not boost")

	// Turn 4: past warmup, should boost
	adj.handleContextInjected(makeEvent("key1"))
	assert.Contains(t, store.boosted, "key1", "post-warmup should boost")
}

func TestRelevanceAdjuster_ActiveBoost(t *testing.T) {
	store := newMockRelevanceStore()
	cfg := activeConfig()
	cfg.WarmupTurns = 0

	adj := NewRelevanceAdjuster(store, cfg, zap.NewNop().Sugar())
	adj.handleContextInjected(makeEvent("key1", "key2"))

	assert.InDelta(t, 0.05, store.boosted["key1"], 0.001)
	assert.InDelta(t, 0.05, store.boosted["key2"], 0.001)
}

func TestRelevanceAdjuster_TurnLevelDedup(t *testing.T) {
	store := newMockRelevanceStore()
	cfg := activeConfig()
	cfg.WarmupTurns = 0

	adj := NewRelevanceAdjuster(store, cfg, zap.NewNop().Sugar())

	// Same key appears twice in one event
	evt := eventbus.ContextInjectedEvent{
		Items: []eventbus.ContextInjectedItem{
			{Layer: "user_knowledge", Key: "dup-key"},
			{Layer: "user_knowledge", Key: "dup-key"},
		},
	}
	adj.handleContextInjected(evt)

	assert.InDelta(t, 0.05, store.boosted["dup-key"], 0.001, "should boost once despite duplicate")
}

func TestRelevanceAdjuster_SkipNonKnowledge(t *testing.T) {
	store := newMockRelevanceStore()
	cfg := activeConfig()
	cfg.WarmupTurns = 0

	adj := NewRelevanceAdjuster(store, cfg, zap.NewNop().Sugar())
	evt := eventbus.ContextInjectedEvent{
		Items: []eventbus.ContextInjectedItem{
			{Layer: "agent_learnings", Key: "learning1"},
			{Layer: "external_knowledge", Key: "ext1"},
			{Layer: "user_knowledge", Key: "kn1"},
		},
	}
	adj.handleContextInjected(evt)

	assert.NotContains(t, store.boosted, "learning1", "learnings should not be boosted")
	assert.NotContains(t, store.boosted, "ext1", "external should not be boosted")
	assert.Contains(t, store.boosted, "kn1", "knowledge should be boosted")
}

func TestRelevanceAdjuster_GlobalDecay(t *testing.T) {
	store := newMockRelevanceStore()
	cfg := activeConfig()
	cfg.WarmupTurns = 0
	cfg.DecayInterval = 5

	adj := NewRelevanceAdjuster(store, cfg, zap.NewNop().Sugar())

	// Turns 1-4: no decay
	for i := 0; i < 4; i++ {
		adj.handleContextInjected(makeEvent("key1"))
	}
	assert.Equal(t, 0, store.decayed, "decay should not fire before interval")

	// Turn 5: decay fires
	adj.handleContextInjected(makeEvent("key1"))
	assert.Equal(t, 1, store.decayed, "decay should fire at interval")
	assert.InDelta(t, 0.01, store.decayDelta, 0.001)
}

func TestRelevanceAdjuster_Rollback(t *testing.T) {
	store := newMockRelevanceStore()
	cfg := activeConfig()
	cfg.WarmupTurns = 0

	adj := NewRelevanceAdjuster(store, cfg, zap.NewNop().Sugar())

	// Active: boost works
	adj.handleContextInjected(makeEvent("key1"))
	assert.Contains(t, store.boosted, "key1")

	// Switch to shadow
	adj.SetMode("shadow")
	store2 := newMockRelevanceStore()
	adj.store = store2
	adj.handleContextInjected(makeEvent("key2"))
	assert.Empty(t, store2.boosted, "shadow mode after rollback should not write")
}

func TestRelevanceAdjuster_DecayBeforeBoost(t *testing.T) {
	store := newMockRelevanceStore()
	cfg := activeConfig()
	cfg.WarmupTurns = 0
	cfg.DecayInterval = 1 // decay every turn

	adj := NewRelevanceAdjuster(store, cfg, zap.NewNop().Sugar())
	adj.handleContextInjected(makeEvent("key1"))

	// Both decay and boost should have fired
	assert.Equal(t, 1, store.decayed, "decay should fire")
	assert.Contains(t, store.boosted, "key1", "boost should fire after decay")
}

func TestRelevanceStore_Reset(t *testing.T) {
	store := newMockRelevanceStore()
	n, err := store.ResetAllRelevanceScores(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.True(t, store.reset)
}
