package runledger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_AppendAndRetrieve(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	// Append two events.
	err := store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "s1", Goal: "test"}),
	})
	require.NoError(t, err)

	err = store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-1",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{StepID: "s1", Goal: "do thing", OwnerAgent: "op", Status: StepStatusPending}},
		}),
	})
	require.NoError(t, err)

	// Retrieve all events.
	events, err := store.GetJournalEvents(ctx, "run-1")
	require.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, int64(1), events[0].Seq)
	assert.Equal(t, int64(2), events[1].Seq)
}

func TestMemoryStore_GetJournalEventsSince(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	for i := 0; i < 5; i++ {
		_ = store.AppendJournalEvent(ctx, JournalEvent{
			RunID:   "run-1",
			Type:    EventNoteWritten,
			Payload: marshalPayload(NoteWrittenPayload{Key: "k", Value: "v"}),
		})
	}

	tail, err := store.GetJournalEventsSince(ctx, "run-1", 3)
	require.NoError(t, err)
	assert.Len(t, tail, 2) // seq 4 and 5
	assert.Equal(t, int64(4), tail[0].Seq)
	assert.Equal(t, int64(5), tail[1].Seq)
}

func TestMemoryStore_MaterializeRunSnapshot(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "s1", Goal: "goal"}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-1",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{
				{StepID: "s1", Goal: "step 1", OwnerAgent: "op", Status: StepStatusPending},
			},
			AcceptanceCriteria: []AcceptanceCriterion{
				{Description: "done", Validator: ValidatorSpec{Type: ValidatorBuildPass}},
			},
		}),
	})

	snap, err := store.MaterializeRunSnapshot(ctx, "run-1")
	require.NoError(t, err)
	assert.Equal(t, "run-1", snap.RunID)
	assert.Equal(t, RunStatusRunning, snap.Status)
	assert.Len(t, snap.Steps, 1)
}

func TestMemoryStore_RecordValidationResult(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{Goal: "test"}),
	})

	err := store.RecordValidationResult(ctx, "run-1", "s1", ValidationResult{
		Passed: true,
		Reason: "build succeeded",
	})
	require.NoError(t, err)

	events, err := store.GetJournalEvents(ctx, "run-1")
	require.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, EventStepValidationPassed, events[1].Type)
}

func TestMemoryStore_SnapshotCaching(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "s1", Goal: "cache test"}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-1",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{StepID: "s1", Goal: "work", OwnerAgent: "op", Status: StepStatusPending}},
		}),
	})

	// First call materializes and caches.
	snap1, err := store.GetRunSnapshot(ctx, "run-1")
	require.NoError(t, err)
	assert.Equal(t, int64(2), snap1.LastJournalSeq)

	// Add another event.
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})

	// Second call should use cache + tail.
	snap2, err := store.GetRunSnapshot(ctx, "run-1")
	require.NoError(t, err)
	assert.Equal(t, int64(3), snap2.LastJournalSeq)
	assert.Equal(t, StepStatusInProgress, snap2.Steps[0].Status)
}

func TestMemoryStore_ListRuns(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	// Create two runs.
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{Goal: "first"}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-2",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{Goal: "second"}),
	})

	runs, err := store.ListRuns(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, runs, 2)
}

func TestMemoryStore_GetJournalEvents_NotFound(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_, err := store.GetJournalEvents(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no journal events")
}
