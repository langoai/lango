package runledger

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/enttest"
	entrunsnapshot "github.com/langoai/lango/internal/ent/runsnapshot"
	"github.com/langoai/lango/internal/storeutil"
)

func newWave39EntStore(t *testing.T) (*EntStore, context.Context) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?_fk=1", filepath.Join(t.TempDir(), "runledger-wave39.db"))
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	return NewEntStore(client), context.Background()
}

func TestWave39EntStoreAppendJournalDefaultsHookAndOrdering(t *testing.T) {
	store, ctx := newWave39EntStore(t)

	var hooked []JournalEvent
	store.SetAppendHook(func(event JournalEvent) {
		hooked = append(hooked, event)
	})

	err := store.AppendJournalEvent(ctx, JournalEvent{
		Type:    EventNoteWritten,
		Payload: marshalPayload(NoteWrittenPayload{Key: "ignored", Value: "missing run"}),
	})
	require.ErrorContains(t, err, "run_id is required")
	require.Empty(t, hooked)

	explicitTime := time.Date(2026, 5, 19, 8, 30, 0, 0, time.UTC)
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-wave39-journal",
		Type:    EventNoteWritten,
		Payload: marshalPayload(NoteWrittenPayload{Key: "first", Value: "defaulted"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		ID:        uuid.NewString(),
		RunID:     "run-wave39-journal",
		Type:      EventProjectionSynced,
		Timestamp: explicitTime,
		Payload:   marshalPayload(ProjectionSyncPayload{Target: "cache", Status: "ok"}),
	}))

	require.Len(t, hooked, 2)
	require.NotEmpty(t, hooked[0].ID)
	_, err = uuid.Parse(hooked[0].ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), hooked[0].Seq)
	assert.False(t, hooked[0].Timestamp.IsZero())
	assert.Equal(t, explicitTime, hooked[1].Timestamp)

	events, err := store.GetJournalEvents(ctx, "run-wave39-journal")
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, []int64{1, 2}, []int64{events[0].Seq, events[1].Seq})
	assert.Equal(t, []JournalEventType{EventNoteWritten, EventProjectionSynced}, []JournalEventType{events[0].Type, events[1].Type})

	tail, err := store.GetJournalEventsSince(ctx, "run-wave39-journal", 1)
	require.NoError(t, err)
	require.Len(t, tail, 1)
	assert.Equal(t, int64(2), tail[0].Seq)
	assert.Equal(t, EventProjectionSynced, tail[0].Type)

	emptyTail, err := store.GetJournalEventsSince(ctx, "run-wave39-journal", 2)
	require.NoError(t, err)
	assert.Empty(t, emptyTail)

	_, err = store.GetJournalEvents(ctx, "missing-run")
	require.ErrorContains(t, err, "no journal events")
}

func TestWave39EntStoreRecordValidationResultPassAndFail(t *testing.T) {
	store, ctx := newWave39EntStore(t)

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-wave39-validation",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "session-wave39", Goal: "validate steps"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-wave39-validation",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{
				{
					StepID:     "step-pass",
					Index:      0,
					Goal:       "pass validation",
					OwnerAgent: "tester",
					Status:     StepStatusPending,
					Validator:  ValidatorSpec{Type: ValidatorTestPass},
					MaxRetries: DefaultMaxRetries,
				},
				{
					StepID:     "step-fail",
					Index:      1,
					Goal:       "fail validation",
					OwnerAgent: "tester",
					Status:     StepStatusPending,
					Validator:  ValidatorSpec{Type: ValidatorBuildPass},
					MaxRetries: DefaultMaxRetries,
				},
			},
		}),
	}))

	require.NoError(t, store.RecordValidationResult(ctx, "run-wave39-validation", "step-pass", ValidationResult{
		Passed: true,
		Reason: "tests passed",
	}))
	require.NoError(t, store.RecordValidationResult(ctx, "run-wave39-validation", "step-fail", ValidationResult{
		Passed:  false,
		Reason:  "build failed",
		Details: map[string]string{"target": "./internal/runledger"},
	}))

	events, err := store.GetJournalEventsSince(ctx, "run-wave39-validation", 2)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, []JournalEventType{EventStepValidationPassed, EventStepValidationFailed}, []JournalEventType{events[0].Type, events[1].Type})

	var passed StepValidationPassedPayload
	require.NoError(t, json.Unmarshal(events[0].Payload, &passed))
	assert.Equal(t, "step-pass", passed.StepID)
	assert.True(t, passed.Result.Passed)

	var failed StepValidationFailedPayload
	require.NoError(t, json.Unmarshal(events[1].Payload, &failed))
	assert.Equal(t, "step-fail", failed.StepID)
	assert.False(t, failed.Result.Passed)
	assert.Equal(t, "./internal/runledger", failed.Result.Details["target"])

	snapshot, err := store.GetRunSnapshot(ctx, "run-wave39-validation")
	require.NoError(t, err)
	require.Len(t, snapshot.Steps, 2)
	assert.Equal(t, StepStatusCompleted, snapshot.Steps[0].Status)
	assert.Equal(t, StepStatusFailed, snapshot.Steps[1].Status)
	assert.Equal(t, "build failed", snapshot.CurrentBlocker)
}

func TestWave39EntStoreCachedSnapshotUpdateReadAndCacheHit(t *testing.T) {
	store, ctx := newWave39EntStore(t)

	snapshot := &RunSnapshot{
		RunID:          "run-wave39-cache",
		SessionKey:     "session-cache",
		Goal:           "persist cache",
		Status:         RunStatusRunning,
		LastJournalSeq: 7,
		Notes:          map[string]string{"phase": "initial"},
		AcceptanceState: []AcceptanceCriterion{{
			Description: "snapshot is stored",
			Validator:   ValidatorSpec{Type: ValidatorArtifactExists, Target: "cache"},
		}},
		Steps: []Step{{
			StepID:     "cache-step",
			Index:      0,
			Goal:       "write snapshot",
			OwnerAgent: "tester",
			Status:     StepStatusVerifyPending,
			Result:     "ready",
			Evidence:   []Evidence{{Type: "file", Content: "internal/runledger/ent_store.go"}},
			Validator:  ValidatorSpec{Type: ValidatorArtifactExists, Target: "internal/runledger/ent_store.go"},
			MaxRetries: DefaultMaxRetries,
		}},
	}

	require.NoError(t, store.UpdateCachedSnapshot(ctx, snapshot))

	snapshot.Goal = "mutated after update"
	snapshot.Steps[0].Goal = "mutated step"
	snapshot.Notes["phase"] = "mutated"

	cached, seq, err := store.GetCachedSnapshot(ctx, "run-wave39-cache")
	require.NoError(t, err)
	require.NotNil(t, cached)
	assert.Equal(t, int64(7), seq)
	assert.Equal(t, "persist cache", cached.Goal)
	assert.Equal(t, "write snapshot", cached.Steps[0].Goal)
	assert.Equal(t, "initial", cached.Notes["phase"])

	reloadedStore := NewEntStore(store.client)
	reloaded, seq, err := reloadedStore.GetCachedSnapshot(ctx, "run-wave39-cache")
	require.NoError(t, err)
	require.NotNil(t, reloaded)
	assert.Equal(t, int64(7), seq)
	assert.Equal(t, "persist cache", reloaded.Goal)
	assert.Equal(t, "cache-step", reloaded.Steps[0].StepID)

	dbChanged := reloaded.DeepCopy()
	dbChanged.Goal = "changed behind cache"
	data, err := storeutil.MarshalField(dbChanged)
	require.NoError(t, err)
	_, err = store.client.RunSnapshot.Update().
		Where(entrunsnapshot.RunIDEQ("run-wave39-cache")).
		SetGoal(dbChanged.Goal).
		SetSnapshotData(string(data)).
		Save(ctx)
	require.NoError(t, err)

	stillCached, seq, err := reloadedStore.GetCachedSnapshot(ctx, "run-wave39-cache")
	require.NoError(t, err)
	require.NotNil(t, stillCached)
	assert.Equal(t, int64(7), seq)
	assert.Equal(t, "persist cache", stillCached.Goal)
}

func TestWave39EntStoreSnapshotErrorBranches(t *testing.T) {
	store, ctx := newWave39EntStore(t)

	err := store.UpdateCachedSnapshot(ctx, &RunSnapshot{
		RunID:            "run-wave39-invalid-json",
		Status:           RunStatusPlanning,
		SourceDescriptor: json.RawMessage(`{`),
	})
	require.ErrorContains(t, err, "marshal snapshot")

	_, err = store.client.RunSnapshot.Create().
		SetRunID("run-wave39-corrupt").
		SetSessionKey("session-corrupt").
		SetStatus(entrunsnapshot.StatusRunning).
		SetGoal("corrupt snapshot").
		SetSnapshotData(`{`).
		SetLastJournalSeq(3).
		Save(ctx)
	require.NoError(t, err)

	corrupt, seq, err := store.GetCachedSnapshot(ctx, "run-wave39-corrupt")
	require.Error(t, err)
	assert.Nil(t, corrupt)
	assert.Zero(t, seq)
	assert.Contains(t, err.Error(), "snapshot run-wave39-corrupt")
}
