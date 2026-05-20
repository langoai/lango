package memory

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/graph"
)

func TestStoreObservationGetPublishesEventsAndGraphHooks(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	bus := eventbus.New()
	var events []eventbus.ContentSavedEvent
	eventbus.SubscribeTyped(bus, func(e eventbus.ContentSavedEvent) {
		events = append(events, e)
	})
	store.SetEventBus(bus)

	var tripleBatches [][]graph.Triple
	store.SetGraphHooks(NewGraphHooks(func(triples []graph.Triple) {
		tripleBatches = append(tripleBatches, triples)
	}, zap.NewNop().Sugar()))

	firstID := uuid.New()
	secondID := uuid.New()
	require.NoError(t, store.SaveObservation(ctx, Observation{
		ID:               firstID,
		SessionKey:       "session-graph",
		Content:          "first observation",
		TokenCount:       3,
		SourceStartIndex: 1,
		SourceEndIndex:   2,
	}))
	require.NoError(t, store.SaveObservation(ctx, Observation{
		ID:         secondID,
		SessionKey: "session-graph",
		Content:    "second observation",
		TokenCount: 4,
	}))

	got, err := store.GetObservation(ctx, firstID)
	require.NoError(t, err)
	assert.Equal(t, firstID, got.ID)
	assert.Equal(t, "session-graph", got.SessionKey)
	assert.Equal(t, "first observation", got.Content)
	assert.Equal(t, 3, got.TokenCount)
	assert.Equal(t, 1, got.SourceStartIndex)
	assert.Equal(t, 2, got.SourceEndIndex)

	require.Len(t, events, 2)
	assert.Equal(t, "observation", events[0].Collection)
	assert.Equal(t, "memory", events[0].Source)
	assert.True(t, events[0].IsNew)
	assert.True(t, events[0].NeedsGraph)
	assert.Equal(t, "session-graph", events[0].Metadata["session_key"])

	require.Len(t, tripleBatches, 2)
	require.Len(t, tripleBatches[0], 1)
	assert.Equal(t, graph.InSession, tripleBatches[0][0].Predicate)
	require.Len(t, tripleBatches[1], 2)
	assert.Equal(t, graph.Follows, tripleBatches[1][1].Predicate)
	assert.Equal(t, "observation:"+firstID.String(), tripleBatches[1][1].Object)
	assert.Equal(t, secondID.String(), events[1].ID)
	assert.Equal(t, "observation", events[1].Collection)
	assert.Equal(t, "second observation", events[1].Content)
	assert.Equal(t, "session-graph", events[1].Metadata["session_key"])
}

func TestStoreReflectionGetDeleteAndGraphHooks(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	obsID := uuid.New()
	require.NoError(t, store.SaveObservation(ctx, Observation{
		ID:         obsID,
		SessionKey: "session-reflect",
		Content:    "source observation",
	}))

	bus := eventbus.New()
	var events []eventbus.ContentSavedEvent
	eventbus.SubscribeTyped(bus, func(e eventbus.ContentSavedEvent) {
		events = append(events, e)
	})
	store.SetEventBus(bus)

	var triples []graph.Triple
	store.SetGraphHooks(NewGraphHooks(func(batch []graph.Triple) {
		triples = append(triples, batch...)
	}, zap.NewNop().Sugar()))

	refID := uuid.New()
	require.NoError(t, store.SaveReflection(ctx, Reflection{
		ID:         refID,
		SessionKey: "session-reflect",
		Content:    "reflection body",
		TokenCount: 9,
		Generation: 3,
	}))

	got, err := store.GetReflection(ctx, refID)
	require.NoError(t, err)
	assert.Equal(t, refID, got.ID)
	assert.Equal(t, "session-reflect", got.SessionKey)
	assert.Equal(t, "reflection body", got.Content)
	assert.Equal(t, 9, got.TokenCount)
	assert.Equal(t, 3, got.Generation)

	require.Len(t, events, 1)
	assert.Equal(t, "reflection", events[0].Collection)
	assert.Equal(t, "reflection body", events[0].Content)
	assert.Equal(t, "session-reflect", events[0].Metadata["session_key"])

	require.GreaterOrEqual(t, len(triples), 2)
	assert.Equal(t, graph.InSession, triples[0].Predicate)
	assert.Equal(t, graph.ReflectsOn, triples[1].Predicate)
	assert.Equal(t, "observation:"+obsID.String(), triples[1].Object)

	require.NoError(t, store.DeleteReflections(ctx, []uuid.UUID{refID}))
	_, err = store.GetReflection(ctx, refID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get reflection")
}

func TestStoreRecentListsReturnOldestFirstWithinLimit(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	base := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)

	for i, content := range []string{"old", "middle", "new"} {
		_, err := store.client.Observation.Create().
			SetSessionKey("recent-session").
			SetContent(content).
			SetCreatedAt(base.Add(time.Duration(i) * time.Minute)).
			Save(ctx)
		require.NoError(t, err)
		_, err = store.client.Reflection.Create().
			SetSessionKey("recent-session").
			SetContent(content + " reflection").
			SetGeneration(i + 1).
			SetCreatedAt(base.Add(time.Duration(i) * time.Minute)).
			Save(ctx)
		require.NoError(t, err)
	}

	observations, err := store.ListRecentObservations(ctx, "recent-session", 2)
	require.NoError(t, err)
	require.Len(t, observations, 2)
	assert.Equal(t, "middle", observations[0].Content)
	assert.Equal(t, "new", observations[1].Content)

	reflections, err := store.ListRecentReflections(ctx, "recent-session", 2)
	require.NoError(t, err)
	require.Len(t, reflections, 2)
	assert.Equal(t, "middle reflection", reflections[0].Content)
	assert.Equal(t, "new reflection", reflections[1].Content)
}
