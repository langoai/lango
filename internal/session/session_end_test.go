package session

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntStore_End_SetsPendingFlag(t *testing.T) {
	store := newTestEntStore(t)

	err := store.Create(&Session{Key: "sess-1"})
	require.NoError(t, err)

	err = store.End("sess-1")
	require.NoError(t, err)

	got, err := store.Get("sess-1")
	require.NoError(t, err)
	assert.True(t, got.EndPending())
}

func TestEntStore_End_Idempotent(t *testing.T) {
	store := newTestEntStore(t)

	err := store.Create(&Session{Key: "sess-1"})
	require.NoError(t, err)

	assert.NoError(t, store.End("sess-1"))
	assert.NoError(t, store.End("sess-1"))
}

func TestEntStore_End_UnknownSession(t *testing.T) {
	store := newTestEntStore(t)

	err := store.End("missing")
	require.Error(t, err)
}

func TestEntStore_End_InvokesProcessor(t *testing.T) {
	store := newTestEntStore(t)
	err := store.Create(&Session{Key: "sess-1"})
	require.NoError(t, err)

	var invoked atomic.Int32
	var gotKey atomic.Value
	store.SetSessionEndProcessor(func(_ context.Context, key string) error {
		invoked.Add(1)
		gotKey.Store(key)
		return nil
	})
	store.SetHardEndTimeout(500 * time.Millisecond)

	err = store.End("sess-1")
	require.NoError(t, err)

	assert.Equal(t, int32(1), invoked.Load())
	assert.Equal(t, "sess-1", gotKey.Load())

	// On processor success the pending flag is cleared.
	got, err := store.Get("sess-1")
	require.NoError(t, err)
	assert.False(t, got.EndPending())
}

func TestEntStore_End_ProcessorTimeout(t *testing.T) {
	store := newTestEntStore(t)
	err := store.Create(&Session{Key: "sess-1"})
	require.NoError(t, err)

	store.SetSessionEndProcessor(func(ctx context.Context, _ string) error {
		<-ctx.Done()
		return ctx.Err()
	})
	store.SetHardEndTimeout(50 * time.Millisecond)

	start := time.Now()
	err = store.End("sess-1")
	elapsed := time.Since(start)
	require.NoError(t, err)

	// Returned within the bound, and pending flag still set so next sweep
	// can recover the job.
	assert.Less(t, elapsed, 500*time.Millisecond)
	got, err := store.Get("sess-1")
	require.NoError(t, err)
	assert.True(t, got.EndPending())
}

func TestEntStore_End_ProcessorFailureKeepsFlag(t *testing.T) {
	store := newTestEntStore(t)
	err := store.Create(&Session{Key: "sess-1"})
	require.NoError(t, err)

	store.SetSessionEndProcessor(func(_ context.Context, _ string) error {
		return errors.New("boom")
	})
	store.SetHardEndTimeout(500 * time.Millisecond)

	err = store.End("sess-1")
	require.NoError(t, err)

	got, err := store.Get("sess-1")
	require.NoError(t, err)
	assert.True(t, got.EndPending(), "processor failure keeps pending flag for retry")
}

func TestEntStore_ListEndPending(t *testing.T) {
	store := newTestEntStore(t)

	require.NoError(t, store.Create(&Session{Key: "sess-1"}))
	require.NoError(t, store.Create(&Session{Key: "sess-2"}))
	require.NoError(t, store.Create(&Session{Key: "sess-3"}))

	require.NoError(t, store.MarkEndPending("sess-1"))
	require.NoError(t, store.MarkEndPending("sess-3"))

	keys, err := store.ListEndPending()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"sess-1", "sess-3"}, keys)
}

func TestEntStore_ClearEndPending(t *testing.T) {
	store := newTestEntStore(t)
	require.NoError(t, store.Create(&Session{Key: "sess-1"}))
	require.NoError(t, store.MarkEndPending("sess-1"))

	require.NoError(t, store.ClearEndPending("sess-1"))

	keys, err := store.ListEndPending()
	require.NoError(t, err)
	assert.Empty(t, keys)
}
