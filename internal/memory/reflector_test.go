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

func newTestReflector(t *testing.T, gen TextGenerator) (*Reflector, *Store) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	logger := zap.NewNop().Sugar()
	store := NewStore(client, logger)
	reflector := NewReflector(gen, store, logger)
	return reflector, store
}

func TestReflect(t *testing.T) {
	t.Run("generates reflection and deletes observations", func(t *testing.T) {
		gen := &mockGenerator{response: "User is building a REST API with Go and PostgreSQL."}
		reflector, store := newTestReflector(t, gen)
		ctx := context.Background()

		// Seed observations.
		obs1 := Observation{
			ID:               uuid.New(),
			SessionKey:       "session-ref-1",
			Content:          "User wants to build a REST API",
			TokenCount:       15,
			SourceStartIndex: 0,
			SourceEndIndex:   5,
		}
		obs2 := Observation{
			ID:               uuid.New(),
			SessionKey:       "session-ref-1",
			Content:          "Decided to use PostgreSQL",
			TokenCount:       10,
			SourceStartIndex: 6,
			SourceEndIndex:   10,
		}
		require.NoError(t, store.SaveObservation(ctx, obs1))
		require.NoError(t, store.SaveObservation(ctx, obs2))

		ref, err := reflector.Reflect(ctx, "session-ref-1")
		require.NoError(t, err)
		require.NotNil(t, ref)

		assert.Equal(t, "User is building a REST API with Go and PostgreSQL.", ref.Content)
		assert.Equal(t, "session-ref-1", ref.SessionKey)
		assert.Equal(t, 1, ref.Generation)
		assert.Greater(t, ref.TokenCount, 0)

		// Verify reflection was saved.
		refs, err := store.ListReflections(ctx, "session-ref-1")
		require.NoError(t, err)
		assert.Len(t, refs, 1)

		// Verify observations were deleted.
		obs, err := store.ListObservations(ctx, "session-ref-1")
		require.NoError(t, err)
		assert.Empty(t, obs)
	})

	t.Run("no observations returns nil", func(t *testing.T) {
		gen := &mockGenerator{response: "should not be called"}
		reflector, _ := newTestReflector(t, gen)
		ctx := context.Background()

		ref, err := reflector.Reflect(ctx, "session-ref-empty")
		require.NoError(t, err)
		assert.Nil(t, ref)
	})
}

func TestReflectOnReflections(t *testing.T) {
	t.Run("generates multi-generation reflection", func(t *testing.T) {
		gen := &mockGenerator{response: "High-level summary of all work done so far."}
		reflector, store := newTestReflector(t, gen)
		ctx := context.Background()

		// Seed gen-1 reflections.
		ref1 := Reflection{
			ID:         uuid.New(),
			SessionKey: "session-meta-1",
			Content:    "User building REST API",
			TokenCount: 10,
			Generation: 1,
		}
		ref2 := Reflection{
			ID:         uuid.New(),
			SessionKey: "session-meta-1",
			Content:    "Switched to GraphQL",
			TokenCount: 8,
			Generation: 1,
		}
		require.NoError(t, store.SaveReflection(ctx, ref1))
		require.NoError(t, store.SaveReflection(ctx, ref2))

		metaRef, err := reflector.ReflectOnReflections(ctx, "session-meta-1")
		require.NoError(t, err)
		require.NotNil(t, metaRef)

		assert.Equal(t, 2, metaRef.Generation)
		assert.Equal(t, "High-level summary of all work done so far.", metaRef.Content)

		// Old reflections deleted, new one saved.
		refs, err := store.ListReflections(ctx, "session-meta-1")
		require.NoError(t, err)
		assert.Len(t, refs, 1)
		assert.Equal(t, 2, refs[0].Generation)
	})

	t.Run("no reflections returns nil", func(t *testing.T) {
		gen := &mockGenerator{response: "should not be called"}
		reflector, _ := newTestReflector(t, gen)
		ctx := context.Background()

		ref, err := reflector.ReflectOnReflections(ctx, "session-meta-empty")
		require.NoError(t, err)
		assert.Nil(t, ref)
	})
}
