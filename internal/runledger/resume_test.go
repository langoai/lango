package runledger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResumeManager_FindCandidates(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	// Create a paused run.
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:     "run-1",
		Type:      EventRunCreated,
		Timestamp: time.Now(),
		Payload: marshalPayload(RunCreatedPayload{
			SessionKey: "session-1",
			Goal:       "paused task",
		}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:     "run-1",
		Type:      EventPlanAttached,
		Timestamp: time.Now(),
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{StepID: "s1", Goal: "work", OwnerAgent: "op", Status: StepStatusPending}},
		}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:     "run-1",
		Type:      EventRunPaused,
		Timestamp: time.Now(),
		Payload:   marshalPayload(RunPausedPayload{Reason: "turn limit"}),
	})

	// Create a completed run (should not be a candidate).
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:     "run-2",
		Type:      EventRunCreated,
		Timestamp: time.Now(),
		Payload: marshalPayload(RunCreatedPayload{
			SessionKey: "session-1",
			Goal:       "done task",
		}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:     "run-2",
		Type:      EventRunCompleted,
		Timestamp: time.Now(),
		Payload:   marshalPayload(RunCompletedPayload{Summary: "done"}),
	})

	rm := NewResumeManager(store, time.Hour)
	candidates, err := rm.FindCandidates(ctx, "session-1")
	require.NoError(t, err)
	assert.Len(t, candidates, 1)
	assert.Equal(t, "run-1", candidates[0].RunID)
	assert.Equal(t, RunStatusPaused, candidates[0].Status)
}

func TestResumeManager_FindCandidates_ExcludesStaleRuns(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	old := time.Now().Add(-2 * time.Hour)
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:     "run-stale",
		Type:      EventRunCreated,
		Timestamp: old,
		Payload: marshalPayload(RunCreatedPayload{
			SessionKey: "session-1",
			Goal:       "stale run",
		}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:     "run-stale",
		Type:      EventPlanAttached,
		Timestamp: old,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{StepID: "s1", Goal: "work", OwnerAgent: "op", Status: StepStatusPending}},
		}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:     "run-stale",
		Type:      EventRunPaused,
		Timestamp: old,
		Payload:   marshalPayload(RunPausedPayload{Reason: "turn limit"}),
	})

	rm := NewResumeManager(store, time.Hour)
	candidates, err := rm.FindCandidates(ctx, "session-1")
	require.NoError(t, err)
	assert.Empty(t, candidates)
}

func TestResumeManager_FindCandidates_MultipleCandidatesSameSession(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	for _, runID := range []string{"run-1", "run-2"} {
		_ = store.AppendJournalEvent(ctx, JournalEvent{
			RunID:   runID,
			Type:    EventRunCreated,
			Payload: marshalPayload(RunCreatedPayload{SessionKey: "session-1", Goal: runID}),
		})
		_ = store.AppendJournalEvent(ctx, JournalEvent{
			RunID: runID,
			Type:  EventPlanAttached,
			Payload: marshalPayload(PlanAttachedPayload{
				Steps: []Step{{StepID: "s1", Goal: "work", OwnerAgent: "op", Status: StepStatusPending}},
			}),
		})
		_ = store.AppendJournalEvent(ctx, JournalEvent{
			RunID:   runID,
			Type:    EventRunPaused,
			Payload: marshalPayload(RunPausedPayload{Reason: "turn limit"}),
		})
	}

	rm := NewResumeManager(store, time.Hour)
	candidates, err := rm.FindCandidates(ctx, "session-1")
	require.NoError(t, err)
	require.Len(t, candidates, 2)
	assert.ElementsMatch(t, []string{"run-1", "run-2"}, []string{candidates[0].RunID, candidates[1].RunID})
}

func TestResumeManager_Resume(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{Goal: "resume test"}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{Steps: []Step{{StepID: "s1", Status: StepStatusPending}}}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunPaused,
		Payload: marshalPayload(RunPausedPayload{Reason: "turn limit"}),
	})

	rm := NewResumeManager(store, time.Hour)

	snap, err := rm.Resume(ctx, "run-1", "user")
	require.NoError(t, err)
	assert.Equal(t, RunStatusRunning, snap.Status)
}

func TestResumeManager_Resume_NotPaused(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{Goal: "active run"}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-1",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{StepID: "s1", Status: StepStatusPending}},
		}),
	})

	rm := NewResumeManager(store, time.Hour)

	_, err := rm.Resume(ctx, "run-1", "user")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRunNotPaused)
}

func TestDetectResumeIntent(t *testing.T) {
	assert.True(t, DetectResumeIntent("계속해줘"))
	assert.True(t, DetectResumeIntent("resume the task"))
	assert.True(t, DetectResumeIntent("Continue from where we stopped"))
	assert.True(t, DetectResumeIntent("이어서 작업"))
	assert.True(t, DetectResumeIntent("마저 해줘"))
	assert.False(t, DetectResumeIntent("start a new project"))
}

func TestBuildStepSummary(t *testing.T) {
	snap := &RunSnapshot{
		CurrentStepID: "s2",
		Steps: []Step{
			{StepID: "s1", Goal: "first", Status: StepStatusCompleted},
			{StepID: "s2", Goal: "second", Status: StepStatusInProgress},
			{StepID: "s3", Goal: "third", Status: StepStatusPending},
		},
	}

	summary := buildStepSummary(snap)
	assert.Contains(t, summary, "current: second")
}
