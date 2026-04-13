package provenance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_SaveAndGet(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	cp := Checkpoint{
		ID:         "cp-1",
		SessionKey: "sess-1",
		RunID:      "run-1",
		Label:      "test checkpoint",
		Trigger:    TriggerManual,
		JournalSeq: 5,
		CreatedAt:  time.Now(),
	}

	require.NoError(t, store.SaveCheckpoint(ctx, cp))

	got, err := store.GetCheckpoint(ctx, "cp-1")
	require.NoError(t, err)
	assert.Equal(t, "cp-1", got.ID)
	assert.Equal(t, "test checkpoint", got.Label)
	assert.Equal(t, TriggerManual, got.Trigger)
	assert.Equal(t, int64(5), got.JournalSeq)
}

func TestMemoryStore_GetNotFound(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	_, err := store.GetCheckpoint(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrCheckpointNotFound)
}

func TestMemoryStore_ListByRun(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	now := time.Now()
	for i := 0; i < 3; i++ {
		require.NoError(t, store.SaveCheckpoint(ctx, Checkpoint{
			ID:         fmt.Sprintf("cp-%d", i),
			RunID:      "run-1",
			SessionKey: "sess-1",
			Label:      fmt.Sprintf("cp %d", i),
			JournalSeq: int64(i + 1),
			CreatedAt:  now.Add(time.Duration(i) * time.Minute),
		}))
	}
	// Different run.
	require.NoError(t, store.SaveCheckpoint(ctx, Checkpoint{
		ID:         "cp-other",
		RunID:      "run-2",
		SessionKey: "sess-1",
		Label:      "other",
		JournalSeq: 1,
		CreatedAt:  now,
	}))

	list, err := store.ListByRun(ctx, "run-1")
	require.NoError(t, err)
	assert.Len(t, list, 3)
	// Ordered by journal_seq asc.
	assert.Equal(t, int64(1), list[0].JournalSeq)
	assert.Equal(t, int64(3), list[2].JournalSeq)
}

func TestMemoryStore_ListBySession(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	now := time.Now()
	for i := 0; i < 5; i++ {
		require.NoError(t, store.SaveCheckpoint(ctx, Checkpoint{
			ID:         fmt.Sprintf("cp-%d", i),
			RunID:      "run-1",
			SessionKey: "sess-1",
			Label:      fmt.Sprintf("cp %d", i),
			JournalSeq: int64(i + 1),
			CreatedAt:  now.Add(time.Duration(i) * time.Minute),
		}))
	}

	// With limit.
	list, err := store.ListBySession(ctx, "sess-1", 3)
	require.NoError(t, err)
	assert.Len(t, list, 3)
	// Ordered by created_at desc.
	assert.True(t, list[0].CreatedAt.After(list[1].CreatedAt))
}

func TestMemoryStore_CountBySession(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		require.NoError(t, store.SaveCheckpoint(ctx, Checkpoint{
			ID:         fmt.Sprintf("cp-%d", i),
			RunID:      "run-1",
			SessionKey: "sess-1",
			Label:      fmt.Sprintf("cp %d", i),
			CreatedAt:  time.Now(),
		}))
	}

	count, err := store.CountBySession(ctx, "sess-1")
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	count, err = store.CountBySession(ctx, "sess-other")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	require.NoError(t, store.SaveCheckpoint(ctx, Checkpoint{
		ID:    "cp-1",
		Label: "test",
	}))

	require.NoError(t, store.DeleteCheckpoint(ctx, "cp-1"))
	_, err := store.GetCheckpoint(ctx, "cp-1")
	assert.ErrorIs(t, err, ErrCheckpointNotFound)
}

func TestMemoryStore_DeleteNotFound(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	err := store.DeleteCheckpoint(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrCheckpointNotFound)
}
