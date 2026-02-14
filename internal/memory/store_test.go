package memory

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/ent/enttest"
	_ "github.com/mattn/go-sqlite3"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	logger := zap.NewNop().Sugar()
	return NewStore(client, logger)
}

func TestSaveAndListObservations(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("save and list", func(t *testing.T) {
		obs := Observation{
			SessionKey:       "session-1",
			Content:          "User wants to build a REST API",
			TokenCount:       15,
			SourceStartIndex: 0,
			SourceEndIndex:   5,
		}
		err := store.SaveObservation(ctx, obs)
		require.NoError(t, err)

		obs2 := Observation{
			SessionKey:       "session-1",
			Content:          "Decided to use Chi router",
			TokenCount:       10,
			SourceStartIndex: 6,
			SourceEndIndex:   10,
		}
		err = store.SaveObservation(ctx, obs2)
		require.NoError(t, err)

		results, err := store.ListObservations(ctx, "session-1")
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "User wants to build a REST API", results[0].Content)
		assert.Equal(t, "Decided to use Chi router", results[1].Content)
		assert.Equal(t, 15, results[0].TokenCount)
		assert.Equal(t, 0, results[0].SourceStartIndex)
		assert.Equal(t, 5, results[0].SourceEndIndex)
	})

	t.Run("list empty session", func(t *testing.T) {
		results, err := store.ListObservations(ctx, "no-such-session")
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("session isolation", func(t *testing.T) {
		obs := Observation{
			SessionKey: "session-2",
			Content:    "Different session content",
			TokenCount: 8,
		}
		err := store.SaveObservation(ctx, obs)
		require.NoError(t, err)

		results, err := store.ListObservations(ctx, "session-2")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Different session content", results[0].Content)
	})
}

func TestDeleteObservations(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Create observations
	obs1 := Observation{
		ID:         uuid.New(),
		SessionKey: "session-del",
		Content:    "First observation",
		TokenCount: 5,
	}
	obs2 := Observation{
		ID:         uuid.New(),
		SessionKey: "session-del",
		Content:    "Second observation",
		TokenCount: 5,
	}
	obs3 := Observation{
		ID:         uuid.New(),
		SessionKey: "session-del",
		Content:    "Third observation",
		TokenCount: 5,
	}
	require.NoError(t, store.SaveObservation(ctx, obs1))
	require.NoError(t, store.SaveObservation(ctx, obs2))
	require.NoError(t, store.SaveObservation(ctx, obs3))

	t.Run("delete by IDs", func(t *testing.T) {
		err := store.DeleteObservations(ctx, []uuid.UUID{obs1.ID, obs2.ID})
		require.NoError(t, err)

		results, err := store.ListObservations(ctx, "session-del")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Third observation", results[0].Content)
	})

	t.Run("delete by session", func(t *testing.T) {
		err := store.DeleteObservationsBySession(ctx, "session-del")
		require.NoError(t, err)

		results, err := store.ListObservations(ctx, "session-del")
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestSaveAndListReflections(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("save and list", func(t *testing.T) {
		ref := Reflection{
			SessionKey: "session-1",
			Content:    "User is building a REST API with Chi router and PostgreSQL",
			TokenCount: 20,
			Generation: 1,
		}
		err := store.SaveReflection(ctx, ref)
		require.NoError(t, err)

		results, err := store.ListReflections(ctx, "session-1")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, ref.Content, results[0].Content)
		assert.Equal(t, 20, results[0].TokenCount)
		assert.Equal(t, 1, results[0].Generation)
	})

	t.Run("list empty session", func(t *testing.T) {
		results, err := store.ListReflections(ctx, "no-such-session")
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("multi-generation reflections", func(t *testing.T) {
		ref2 := Reflection{
			SessionKey: "session-1",
			Content:    "High-level summary of all work done",
			TokenCount: 10,
			Generation: 2,
		}
		err := store.SaveReflection(ctx, ref2)
		require.NoError(t, err)

		results, err := store.ListReflections(ctx, "session-1")
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, 1, results[0].Generation)
		assert.Equal(t, 2, results[1].Generation)
	})
}

func TestDeleteReflectionsBySession(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Create reflections in different sessions
	ref1 := Reflection{
		SessionKey: "session-ref-del",
		Content:    "First reflection",
		TokenCount: 10,
		Generation: 1,
	}
	ref2 := Reflection{
		SessionKey: "session-ref-del",
		Content:    "Second reflection",
		TokenCount: 15,
		Generation: 2,
	}
	ref3 := Reflection{
		SessionKey: "session-ref-other",
		Content:    "Other session reflection",
		TokenCount: 8,
		Generation: 1,
	}
	require.NoError(t, store.SaveReflection(ctx, ref1))
	require.NoError(t, store.SaveReflection(ctx, ref2))
	require.NoError(t, store.SaveReflection(ctx, ref3))

	t.Run("delete by session", func(t *testing.T) {
		err := store.DeleteReflectionsBySession(ctx, "session-ref-del")
		require.NoError(t, err)

		results, err := store.ListReflections(ctx, "session-ref-del")
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("other session unaffected", func(t *testing.T) {
		results, err := store.ListReflections(ctx, "session-ref-other")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Other session reflection", results[0].Content)
	})

	t.Run("delete empty session", func(t *testing.T) {
		err := store.DeleteReflectionsBySession(ctx, "no-such-session")
		require.NoError(t, err)
	})
}
