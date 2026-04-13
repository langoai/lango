package runledger

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/enttest"
)

func TestEntStore_JournalAndSnapshotRoundTrip(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	store := NewEntStore(client)
	ctx := context.Background()

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "s1", Goal: "goal"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-1",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{
				StepID:     "step-1",
				Goal:       "do work",
				OwnerAgent: "operator",
				Status:     StepStatusPending,
				Validator:  ValidatorSpec{Type: ValidatorBuildPass},
				MaxRetries: DefaultMaxRetries,
			}},
		}),
	}))

	snap, err := store.GetRunSnapshot(ctx, "run-1")
	require.NoError(t, err)
	require.NoError(t, store.UpdateCachedSnapshot(ctx, snap))

	reloaded, err := store.GetRunSnapshot(ctx, "run-1")
	require.NoError(t, err)
	assert.Equal(t, "run-1", reloaded.RunID)
	assert.Equal(t, "goal", reloaded.Goal)
	require.Len(t, reloaded.Steps, 1)
	assert.Equal(t, "step-1", reloaded.Steps[0].StepID)
}

func TestEntStore_ListRunsUsesPersistentSnapshots(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	store := NewEntStore(client)
	ctx := context.Background()

	for _, runID := range []string{"run-a", "run-b"} {
		require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
			RunID:   runID,
			Type:    EventRunCreated,
			Payload: marshalPayload(RunCreatedPayload{SessionKey: "s", Goal: runID}),
		}))
		require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
			RunID: runID,
			Type:  EventPlanAttached,
			Payload: marshalPayload(PlanAttachedPayload{
				Steps: []Step{{
					StepID:     "s1",
					Goal:       "g",
					OwnerAgent: "operator",
					Status:     StepStatusPending,
					Validator:  ValidatorSpec{Type: ValidatorBuildPass},
					MaxRetries: DefaultMaxRetries,
				}},
			}),
		}))
		snap, err := store.GetRunSnapshot(ctx, runID)
		require.NoError(t, err)
		require.NoError(t, store.UpdateCachedSnapshot(ctx, snap))
	}

	runs, err := store.ListRuns(ctx, 10)
	require.NoError(t, err)
	require.Len(t, runs, 2)
	assert.ElementsMatch(t, []string{"run-a", "run-b"}, []string{runs[0].RunID, runs[1].RunID})
}

func TestEntStore_PruneOldRuns(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	store := NewEntStore(client)
	ctx := context.Background()

	seedRun := func(runID string, status RunStatus) {
		require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
			RunID:   runID,
			Type:    EventRunCreated,
			Payload: marshalPayload(RunCreatedPayload{SessionKey: "s", Goal: runID}),
		}))
		require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
			RunID: runID,
			Type:  EventPlanAttached,
			Payload: marshalPayload(PlanAttachedPayload{
				Steps: []Step{{
					StepID:     "s1",
					Goal:       "g",
					OwnerAgent: "operator",
					Status:     StepStatusPending,
					Validator:  ValidatorSpec{Type: ValidatorBuildPass},
					MaxRetries: DefaultMaxRetries,
				}},
			}),
		}))
		snap, err := store.GetRunSnapshot(ctx, runID)
		require.NoError(t, err)
		require.NoError(t, store.UpdateCachedSnapshot(ctx, snap))
		switch status {
		case RunStatusCompleted:
			require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
				RunID:   runID,
				Type:    EventRunCompleted,
				Payload: marshalPayload(RunCompletedPayload{Summary: "done"}),
			}))
		case RunStatusFailed:
			require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
				RunID:   runID,
				Type:    EventRunFailed,
				Payload: marshalPayload(RunFailedPayload{Reason: "failed"}),
			}))
		}
		updated, err := store.GetRunSnapshot(ctx, runID)
		require.NoError(t, err)
		require.NoError(t, store.UpdateCachedSnapshot(ctx, updated))
	}

	seedRun("run-1", RunStatusCompleted)
	seedRun("run-2", RunStatusFailed)
	seedRun("run-3", RunStatusRunning)

	require.NoError(t, store.PruneOldRuns(ctx, 2))

	runs, err := store.ListRuns(ctx, 10)
	require.NoError(t, err)
	require.Len(t, runs, 2)
	assert.ElementsMatch(t, []string{"run-2", "run-3"}, []string{runs[0].RunID, runs[1].RunID})
}

func TestEntStore_GetRunSnapshot_ReturnsIndependentCopy(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	store := NewEntStore(client)
	ctx := context.Background()

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-copy",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "s", Goal: "copy"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-copy",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{
				StepID:     "s1",
				Goal:       "work",
				OwnerAgent: "operator",
				Status:     StepStatusPending,
				Validator: ValidatorSpec{
					Type:   ValidatorBuildPass,
					Params: map[string]string{"mode": "fast"},
				},
			}},
			AcceptanceCriteria: []AcceptanceCriterion{{
				Description: "build",
				Validator:   ValidatorSpec{Type: ValidatorBuildPass},
			}},
		}),
	}))

	snap, err := store.GetRunSnapshot(ctx, "run-copy")
	require.NoError(t, err)
	require.NoError(t, store.UpdateCachedSnapshot(ctx, snap))

	snap.Steps[0].Goal = "changed"
	snap.Steps[0].Validator.Params["mode"] = "slow"
	snap.AcceptanceState[0].Met = true

	again, err := store.GetRunSnapshot(ctx, "run-copy")
	require.NoError(t, err)
	assert.Equal(t, "work", again.Steps[0].Goal)
	assert.Equal(t, "fast", again.Steps[0].Validator.Params["mode"])
	assert.False(t, again.AcceptanceState[0].Met)
}

func TestEntStore_GetRunSnapshot_RaceRegression(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	store := NewEntStore(client)
	ctx := context.Background()

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-race",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "s", Goal: "race"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-race",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{
				StepID:     "s1",
				Goal:       "work",
				OwnerAgent: "operator",
				Status:     StepStatusPending,
				Validator:  ValidatorSpec{Type: ValidatorBuildPass},
			}},
		}),
	}))

	initial, err := store.GetRunSnapshot(ctx, "run-race")
	require.NoError(t, err)
	require.NoError(t, store.UpdateCachedSnapshot(ctx, initial))

	done := make(chan struct{})
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				_ = initial.LastJournalSeq
				if len(initial.Steps) > 0 {
					_ = initial.Steps[0].Status
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 30; i++ {
			if err := store.AppendJournalEvent(ctx, JournalEvent{
				RunID:   "run-race",
				Type:    EventNoteWritten,
				Payload: marshalPayload(NoteWrittenPayload{Key: "k", Value: "v"}),
			}); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			if _, err := store.GetRunSnapshot(ctx, "run-race"); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)
	close(done)
	wg.Wait()
	select {
	case err := <-errCh:
		require.NoError(t, err)
	default:
	}
}

func TestAppendJournalEvent_ConcurrentSameRun(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?_fk=1", filepath.Join(t.TempDir(), "ent-concurrent.db"))
	client := enttest.Open(t, "sqlite3", dsn)
	defer client.Close()

	store := NewEntStore(client)
	ctx := context.Background()

	const workers = 16
	errCh := make(chan error, workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := store.AppendJournalEvent(ctx, JournalEvent{
				RunID:   "run-same",
				Type:    EventNoteWritten,
				Payload: marshalPayload(NoteWrittenPayload{Key: "k", Value: "v"}),
			})
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	events, err := store.GetJournalEvents(ctx, "run-same")
	require.NoError(t, err)
	require.Len(t, events, workers)
	for i, event := range events {
		assert.Equal(t, int64(i+1), event.Seq)
	}
}

func TestGetCachedSnapshot_ConcurrentDifferentRuns(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?_fk=1", filepath.Join(t.TempDir(), "ent-locks.db"))
	client := enttest.Open(t, "sqlite3", dsn)
	defer client.Close()

	store := NewEntStore(client)
	store.cache.Store("run-2", (&RunSnapshot{
		RunID:          "run-2",
		LastJournalSeq: 3,
	}).DeepCopy())

	lock := store.runLock("run-1")
	lock.Lock()
	defer lock.Unlock()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _, _ = store.GetCachedSnapshot(context.Background(), "run-2")
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("GetCachedSnapshot for run-2 blocked on run-1 lock")
	}
}

func TestEntStore_AppendHookReceivesSeq(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	var received []int64
	store := NewEntStore(client, WithAppendHook(func(event JournalEvent) {
		received = append(received, event.Seq)
	}))
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
			RunID:   "run-seq",
			Type:    EventNoteWritten,
			Payload: marshalPayload(NoteWrittenPayload{Key: "k", Value: "v"}),
		}))
	}

	require.Len(t, received, 3)
	assert.Equal(t, int64(1), received[0])
	assert.Equal(t, int64(2), received[1])
	assert.Equal(t, int64(3), received[2])
}

func TestEntStore_SetAppendHook(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	store := NewEntStore(client)
	ctx := context.Background()

	var hookCalled bool
	store.SetAppendHook(func(_ JournalEvent) {
		hookCalled = true
	})

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-hook",
		Type:    EventNoteWritten,
		Payload: marshalPayload(NoteWrittenPayload{Key: "k", Value: "v"}),
	}))

	assert.True(t, hookCalled)
}

func TestEntStore_SetAppendHook_Chaining(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	var calls []string
	store := NewEntStore(client, WithAppendHook(func(_ JournalEvent) {
		calls = append(calls, "first")
	}))
	store.SetAppendHook(func(_ JournalEvent) {
		calls = append(calls, "second")
	})

	ctx := context.Background()
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-chain",
		Type:    EventNoteWritten,
		Payload: marshalPayload(NoteWrittenPayload{Key: "k", Value: "v"}),
	}))

	assert.Equal(t, []string{"first", "second"}, calls)
}
