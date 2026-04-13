package provenance

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/enttest"
)

func newTestEntCheckpointStore(t *testing.T) *EntCheckpointStore {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	return NewEntCheckpointStore(client)
}

func TestEntCheckpointStore_SaveAndGet(t *testing.T) {
	store := newTestEntCheckpointStore(t)
	ctx := context.Background()

	cp := Checkpoint{
		ID:         "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		SessionKey: "sess-1",
		RunID:      "run-1",
		Label:      "test checkpoint",
		Trigger:    TriggerManual,
		JournalSeq: 5,
		GitRef:     "abc123",
		Metadata:   map[string]string{"key": "val"},
		CreatedAt:  time.Now().Truncate(time.Second),
	}

	require.NoError(t, store.SaveCheckpoint(ctx, cp))

	got, err := store.GetCheckpoint(ctx, cp.ID)
	require.NoError(t, err)
	assert.Equal(t, cp.ID, got.ID)
	assert.Equal(t, cp.SessionKey, got.SessionKey)
	assert.Equal(t, cp.RunID, got.RunID)
	assert.Equal(t, cp.Label, got.Label)
	assert.Equal(t, cp.Trigger, got.Trigger)
	assert.Equal(t, cp.JournalSeq, got.JournalSeq)
	assert.Equal(t, cp.GitRef, got.GitRef)
	assert.Equal(t, "val", got.Metadata["key"])
}

func TestEntCheckpointStore_GetNotFound(t *testing.T) {
	store := newTestEntCheckpointStore(t)
	ctx := context.Background()

	_, err := store.GetCheckpoint(ctx, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22")
	assert.ErrorIs(t, err, ErrCheckpointNotFound)
}

func TestEntCheckpointStore_GetInvalidUUID(t *testing.T) {
	store := newTestEntCheckpointStore(t)
	ctx := context.Background()

	_, err := store.GetCheckpoint(ctx, "invalid-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse checkpoint id")
}

func TestEntCheckpointStore_ListByRun(t *testing.T) {
	store := newTestEntCheckpointStore(t)
	ctx := context.Background()

	for i, id := range []string{
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13",
	} {
		require.NoError(t, store.SaveCheckpoint(ctx, Checkpoint{
			ID:         id,
			SessionKey: "sess-1",
			RunID:      "run-1",
			Label:      "cp",
			Trigger:    TriggerManual,
			JournalSeq: int64(i + 1),
			CreatedAt:  time.Now(),
		}))
	}
	// Different run.
	require.NoError(t, store.SaveCheckpoint(ctx, Checkpoint{
		ID:         "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14",
		SessionKey: "sess-1",
		RunID:      "run-2",
		Label:      "other",
		Trigger:    TriggerManual,
		JournalSeq: 1,
		CreatedAt:  time.Now(),
	}))

	list, err := store.ListByRun(ctx, "run-1")
	require.NoError(t, err)
	require.Len(t, list, 3)
	// Should be ordered by journal_seq asc.
	assert.Equal(t, int64(1), list[0].JournalSeq)
	assert.Equal(t, int64(2), list[1].JournalSeq)
	assert.Equal(t, int64(3), list[2].JournalSeq)
}

func TestEntCheckpointStore_ListBySession(t *testing.T) {
	store := newTestEntCheckpointStore(t)
	ctx := context.Background()

	now := time.Now()
	for i, id := range []string{
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13",
	} {
		require.NoError(t, store.SaveCheckpoint(ctx, Checkpoint{
			ID:         id,
			SessionKey: "sess-1",
			RunID:      "run-1",
			Label:      "cp",
			Trigger:    TriggerManual,
			JournalSeq: int64(i + 1),
			CreatedAt:  now.Add(time.Duration(i) * time.Second),
		}))
	}

	// Limit=2, ordered by created_at desc.
	list, err := store.ListBySession(ctx, "sess-1", 2)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(3), list[0].JournalSeq) // newest first
	assert.Equal(t, int64(2), list[1].JournalSeq)
}

func TestEntCheckpointStore_CountBySession(t *testing.T) {
	store := newTestEntCheckpointStore(t)
	ctx := context.Background()

	for i, id := range []string{
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
	} {
		require.NoError(t, store.SaveCheckpoint(ctx, Checkpoint{
			ID:         id,
			SessionKey: "sess-1",
			RunID:      "run-1",
			Label:      "cp",
			Trigger:    TriggerManual,
			JournalSeq: int64(i + 1),
			CreatedAt:  time.Now(),
		}))
	}

	count, err := store.CountBySession(ctx, "sess-1")
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	count, err = store.CountBySession(ctx, "sess-nonexistent")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestEntCheckpointStore_DeleteCheckpoint(t *testing.T) {
	store := newTestEntCheckpointStore(t)
	ctx := context.Background()

	cp := Checkpoint{
		ID:         "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		SessionKey: "sess-1",
		RunID:      "run-1",
		Label:      "to delete",
		Trigger:    TriggerManual,
		JournalSeq: 1,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, store.SaveCheckpoint(ctx, cp))

	require.NoError(t, store.DeleteCheckpoint(ctx, cp.ID))

	_, err := store.GetCheckpoint(ctx, cp.ID)
	assert.ErrorIs(t, err, ErrCheckpointNotFound)
}

func TestEntCheckpointStore_DeleteNotFound(t *testing.T) {
	store := newTestEntCheckpointStore(t)
	ctx := context.Background()

	err := store.DeleteCheckpoint(ctx, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a99")
	assert.ErrorIs(t, err, ErrCheckpointNotFound)
}

func TestEntCheckpointStore_RoundTrip(t *testing.T) {
	store := newTestEntCheckpointStore(t)
	ctx := context.Background()

	// Create -> List -> Show round-trip.
	for i, id := range []string{
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
	} {
		require.NoError(t, store.SaveCheckpoint(ctx, Checkpoint{
			ID:         id,
			SessionKey: "sess-rt",
			RunID:      "run-rt",
			Label:      "round-trip",
			Trigger:    TriggerStepComplete,
			JournalSeq: int64(i + 1),
			CreatedAt:  time.Now(),
		}))
	}

	list, err := store.ListByRun(ctx, "run-rt")
	require.NoError(t, err)
	require.Len(t, list, 2)

	for _, cp := range list {
		got, getErr := store.GetCheckpoint(ctx, cp.ID)
		require.NoError(t, getErr)
		assert.Equal(t, cp.Label, got.Label)
		assert.Equal(t, cp.JournalSeq, got.JournalSeq)
	}
}
