package runledger

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/enttest"
	entrunjournal "github.com/langoai/lango/internal/ent/runjournal"
	entrunsnapshot "github.com/langoai/lango/internal/ent/runsnapshot"
	entrunstep "github.com/langoai/lango/internal/ent/runstep"
)

func newEntStoreResidualBranchesStore(t *testing.T) (*EntStore, context.Context) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?_fk=1", filepath.Join(t.TempDir(), "runledger-ent-store-residual.db"))
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	return NewEntStore(client), context.Background()
}

func appendResidualRunCreatedAndPlan(t *testing.T, ctx context.Context, store *EntStore, runID string) {
	t.Helper()

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{
			SessionKey: "session-" + runID,
			Goal:       "cover ent store residual branches",
		}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{
				StepID:     "step-1",
				Index:      0,
				Goal:       "exercise storage",
				OwnerAgent: "tester",
				Status:     StepStatusPending,
				Validator:  ValidatorSpec{Type: ValidatorTestPass, Target: "./internal/runledger"},
				MaxRetries: DefaultMaxRetries,
			}},
		}),
	}))
}

func TestEntStoreResidualBranchesGetJournalEventsSinceAndMaterializeWithoutCache(t *testing.T) {
	store, ctx := newEntStoreResidualBranchesStore(t)
	const runID = "run-ent-residual-journal"

	appendResidualRunCreatedAndPlan(t, ctx, store, runID)
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventNoteWritten,
		Payload: marshalPayload(NoteWrittenPayload{Key: "note", Value: "covered"}),
	}))

	tail, err := store.GetJournalEventsSince(ctx, runID, 1)
	require.NoError(t, err)
	require.Len(t, tail, 2)
	assert.Equal(t, int64(2), tail[0].Seq)
	assert.Equal(t, EventPlanAttached, tail[0].Type)
	assert.Equal(t, int64(3), tail[1].Seq)
	assert.Equal(t, EventNoteWritten, tail[1].Type)

	emptyTail, err := store.GetJournalEventsSince(ctx, runID, 3)
	require.NoError(t, err)
	assert.Empty(t, emptyTail)

	snapshot, err := store.GetRunSnapshot(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, runID, snapshot.RunID)
	assert.Equal(t, int64(3), snapshot.LastJournalSeq)
	assert.Equal(t, "covered", snapshot.Notes["note"])

	count, err := store.client.RunSnapshot.Query().
		Where(entrunsnapshot.RunIDEQ(runID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestEntStoreResidualBranchesGetCachedSnapshotMalformedAndRunIDBackfill(t *testing.T) {
	store, ctx := newEntStoreResidualBranchesStore(t)

	_, err := store.client.RunSnapshot.Create().
		SetRunID("run-ent-residual-malformed-cache").
		SetSessionKey("session-corrupt").
		SetStatus(entrunsnapshot.StatusRunning).
		SetGoal("corrupt cache row").
		SetSnapshotData(`{`).
		SetLastJournalSeq(7).
		Save(ctx)
	require.NoError(t, err)

	corrupt, seq, err := store.GetCachedSnapshot(ctx, "run-ent-residual-malformed-cache")
	require.Error(t, err)
	assert.Nil(t, corrupt)
	assert.Zero(t, seq)
	assert.Contains(t, err.Error(), "snapshot run-ent-residual-malformed-cache")

	_, err = store.client.RunSnapshot.Create().
		SetRunID("run-ent-residual-backfill").
		SetSessionKey("session-backfill").
		SetStatus(entrunsnapshot.StatusRunning).
		SetGoal("backfill run id").
		SetSnapshotData(`{"session_key":"session-backfill","goal":"backfill run id","status":"running","notes":{},"last_journal_seq":4}`).
		SetLastJournalSeq(4).
		Save(ctx)
	require.NoError(t, err)

	backfilled, seq, err := store.GetCachedSnapshot(ctx, "run-ent-residual-backfill")
	require.NoError(t, err)
	require.NotNil(t, backfilled)
	assert.Equal(t, "run-ent-residual-backfill", backfilled.RunID)
	assert.Equal(t, int64(4), backfilled.LastJournalSeq)
	assert.Equal(t, int64(4), seq)
}

func TestEntStoreResidualBranchesUpdateCachedSnapshotUpdatesRowStepsAndCopiesCache(t *testing.T) {
	store, ctx := newEntStoreResidualBranchesStore(t)
	const runID = "run-ent-residual-update-cache"

	initial := &RunSnapshot{
		RunID:          runID,
		SessionKey:     "session-initial",
		Goal:           "initial goal",
		Status:         RunStatusRunning,
		LastJournalSeq: 1,
		Steps: []Step{{
			StepID:     "old-step",
			Index:      0,
			Goal:       "old step",
			OwnerAgent: "tester",
			Status:     StepStatusPending,
			Validator:  ValidatorSpec{Type: ValidatorBuildPass},
			MaxRetries: DefaultMaxRetries,
		}},
	}
	require.NoError(t, store.UpdateCachedSnapshot(ctx, initial))

	updated := &RunSnapshot{
		RunID:          runID,
		SessionKey:     "session-updated",
		Goal:           "updated goal",
		Status:         RunStatusCompleted,
		LastJournalSeq: 2,
		Steps: []Step{{
			StepID:     "new-step",
			Index:      0,
			Goal:       "new step",
			OwnerAgent: "reviewer",
			Status:     StepStatusCompleted,
			Result:     "done",
			Evidence:   []Evidence{{Type: "file", Content: "internal/runledger/ent_store.go"}},
			Validator:  ValidatorSpec{Type: ValidatorTestPass, Target: "./internal/runledger"},
			MaxRetries: DefaultMaxRetries,
		}},
	}
	require.NoError(t, store.UpdateCachedSnapshot(ctx, updated))

	snapshotCount, err := store.client.RunSnapshot.Query().
		Where(entrunsnapshot.RunIDEQ(runID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, snapshotCount)

	row, err := store.client.RunSnapshot.Query().
		Where(entrunsnapshot.RunIDEQ(runID)).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, "session-updated", row.SessionKey)
	assert.Equal(t, "updated goal", row.Goal)
	assert.Equal(t, entrunsnapshot.StatusCompleted, row.Status)
	assert.Equal(t, int64(2), row.LastJournalSeq)

	steps, err := store.client.RunStep.Query().
		Where(entrunstep.RunIDEQ(runID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, steps, 1)
	assert.Equal(t, "new-step", steps[0].StepID)
	assert.Equal(t, entrunstep.StatusCompleted, steps[0].Status)
	assert.Equal(t, "done", steps[0].Result)

	updated.Goal = "mutated after update"
	updated.Steps[0].Goal = "mutated step"
	updated.Steps[0].Evidence[0].Content = "mutated"

	cached, err := store.GetRunSnapshot(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, "updated goal", cached.Goal)
	require.Len(t, cached.Steps, 1)
	assert.Equal(t, "new step", cached.Steps[0].Goal)
	assert.Equal(t, "internal/runledger/ent_store.go", cached.Steps[0].Evidence[0].Content)
}

func TestEntStoreResidualBranchesPruneDeletesTerminalRowsAndPreservesActiveRuns(t *testing.T) {
	store, ctx := newEntStoreResidualBranchesStore(t)

	seed := func(runID string, status RunStatus) {
		appendResidualRunCreatedAndPlan(t, ctx, store, runID)
		snapshot, err := store.GetRunSnapshot(ctx, runID)
		require.NoError(t, err)
		snapshot.Status = status
		require.NoError(t, store.UpdateCachedSnapshot(ctx, snapshot))
	}

	seed("run-ent-residual-prune-completed", RunStatusCompleted)
	seed("run-ent-residual-prune-active", RunStatusRunning)
	seed("run-ent-residual-prune-failed", RunStatusFailed)

	require.NoError(t, store.PruneOldRuns(ctx, 1))

	assertResidualRunDeleted(t, ctx, store, "run-ent-residual-prune-completed")
	assertResidualRunPresent(t, ctx, store, "run-ent-residual-prune-active")
	assertResidualRunDeleted(t, ctx, store, "run-ent-residual-prune-failed")

	cached, seq, err := store.GetCachedSnapshot(ctx, "run-ent-residual-prune-completed")
	require.NoError(t, err)
	assert.Nil(t, cached)
	assert.Zero(t, seq)
}

func TestShouldRetryAppendJournalErrorStringBranches(t *testing.T) {
	assert.True(t, shouldRetryAppendJournalError(errors.New("create run_journal: database table is locked")))
	assert.True(t, shouldRetryAppendJournalError(errors.New("commit journal event: DATABASE IS LOCKED")))
	assert.False(t, shouldRetryAppendJournalError(errors.New("create run_journal: disk I/O error")))
}

func TestShouldRetryAppendJournalErrorEntConstraintBranch(t *testing.T) {
	store, ctx := newEntStoreResidualBranchesStore(t)

	_, err := store.client.RunSnapshot.Create().
		SetRunID("run-ent-residual-constraint").
		SetSessionKey("session-constraint").
		SetStatus(entrunsnapshot.StatusRunning).
		SetGoal("constraint branch").
		SetSnapshotData(`{"run_id":"run-ent-residual-constraint","session_key":"session-constraint","status":"running","notes":{}}`).
		SetLastJournalSeq(1).
		Save(ctx)
	require.NoError(t, err)

	_, err = store.client.RunSnapshot.Create().
		SetRunID("run-ent-residual-constraint").
		SetSessionKey("session-constraint").
		SetStatus(entrunsnapshot.StatusRunning).
		SetGoal("constraint branch duplicate").
		SetSnapshotData(`{"run_id":"run-ent-residual-constraint","session_key":"session-constraint","status":"running","notes":{}}`).
		SetLastJournalSeq(2).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, shouldRetryAppendJournalError(err))
}

func TestEntStoreResidualBranchesClosedClientReportsStorageErrors(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?_fk=1", filepath.Join(t.TempDir(), "runledger-ent-store-closed.db"))
	client := enttest.Open(t, "sqlite3", dsn)
	store := NewEntStore(client)
	ctx := context.Background()
	require.NoError(t, client.Close())

	err := store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-ent-residual-closed",
		Type:    EventNoteWritten,
		Payload: marshalPayload(NoteWrittenPayload{Key: "closed", Value: "client"}),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "begin tx")

	events, err := store.GetJournalEvents(ctx, "run-ent-residual-closed")
	require.Error(t, err)
	assert.Nil(t, events)
	assert.Contains(t, err.Error(), "query journal events")

	events, err = store.GetJournalEventsSince(ctx, "run-ent-residual-closed", 1)
	require.Error(t, err)
	assert.Nil(t, events)
	assert.Contains(t, err.Error(), "query journal events since 1")

	snapshot, err := store.MaterializeRunSnapshot(ctx, "run-ent-residual-closed")
	require.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), "query journal events")

	cached, seq, err := store.GetCachedSnapshot(ctx, "run-ent-residual-closed")
	require.Error(t, err)
	assert.Nil(t, cached)
	assert.Zero(t, seq)
	assert.Contains(t, err.Error(), "query cached snapshot")

	err = store.UpdateCachedSnapshot(ctx, &RunSnapshot{
		RunID:          "run-ent-residual-closed",
		SessionKey:     "session-closed",
		Status:         RunStatusRunning,
		LastJournalSeq: 1,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "begin snapshot tx")

	runs, err := store.ListRuns(ctx, 10)
	require.Error(t, err)
	assert.Nil(t, runs)
	assert.Contains(t, err.Error(), "list run snapshots")

	sessionRuns, err := store.ListRunSummariesBySession(ctx, "session-closed", 10)
	require.Error(t, err)
	assert.Nil(t, sessionRuns)
	assert.Contains(t, err.Error(), "list run summaries by session")

	maxSeq, err := store.MaxJournalSeqForSession(ctx, "session-closed")
	require.Error(t, err)
	assert.Zero(t, maxSeq)
	assert.Contains(t, err.Error(), "query max journal seq for session")

	err = store.PruneOldRuns(ctx, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "count run snapshots")
}

func TestEntStoreResidualBranchesGetRunSnapshotCachedTailErrors(t *testing.T) {
	store, ctx := newEntStoreResidualBranchesStore(t)
	const runID = "run-ent-residual-tail-error"

	appendResidualRunCreatedAndPlan(t, ctx, store, runID)
	initial, err := store.GetRunSnapshot(ctx, runID)
	require.NoError(t, err)
	require.Equal(t, int64(2), initial.LastJournalSeq)

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventStepStarted,
		Payload: []byte(`{`),
	}))

	snapshot, err := store.GetRunSnapshot(ctx, runID)
	require.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), "apply tail")
	assert.Contains(t, err.Error(), "unmarshal step_started")
}

func TestEntStoreResidualBranchesGetRunSnapshotCachedTailQueryError(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?_fk=1", filepath.Join(t.TempDir(), "runledger-ent-store-tail-query-closed.db"))
	client := enttest.Open(t, "sqlite3", dsn)
	store := NewEntStore(client)
	ctx := context.Background()

	store.cache.Store("run-ent-residual-tail-query-closed", &RunSnapshot{
		RunID:          "run-ent-residual-tail-query-closed",
		SessionKey:     "session-tail-query-closed",
		Status:         RunStatusRunning,
		LastJournalSeq: 5,
	})
	require.NoError(t, client.Close())

	snapshot, err := store.GetRunSnapshot(ctx, "run-ent-residual-tail-query-closed")
	require.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), "query journal events since 5")
}

func TestEntStoreResidualBranchesMaterializeMissingRunReportsNoEvents(t *testing.T) {
	store, ctx := newEntStoreResidualBranchesStore(t)

	snapshot, err := store.MaterializeRunSnapshot(ctx, "run-ent-residual-missing")
	require.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), `run "run-ent-residual-missing": no journal events`)
}

func TestEntStoreResidualBranchesUpdateCachedSnapshotValidationErrorsRollback(t *testing.T) {
	store, ctx := newEntStoreResidualBranchesStore(t)

	err := store.UpdateCachedSnapshot(ctx, &RunSnapshot{
		RunID:          "run-ent-residual-invalid-status",
		SessionKey:     "session-invalid-status",
		Status:         RunStatus("invalid"),
		LastJournalSeq: 1,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upsert snapshot")
	assert.Contains(t, err.Error(), "invalid enum value for status")

	err = store.UpdateCachedSnapshot(ctx, &RunSnapshot{
		RunID:          "run-ent-residual-invalid-step-status",
		SessionKey:     "session-invalid-step-status",
		Status:         RunStatusRunning,
		LastJournalSeq: 1,
		Steps: []Step{{
			StepID:     "step-invalid",
			Goal:       "invalid step status rolls back",
			OwnerAgent: "tester",
			Status:     StepStatus("invalid"),
			Validator:  ValidatorSpec{Type: ValidatorBuildPass},
			MaxRetries: DefaultMaxRetries,
		}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `create run step "step-invalid"`)
	assert.Contains(t, err.Error(), "invalid enum value for status")

	count, err := store.client.RunSnapshot.Query().
		Where(entrunsnapshot.RunIDEQ("run-ent-residual-invalid-step-status")).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func assertResidualRunDeleted(t *testing.T, ctx context.Context, store *EntStore, runID string) {
	t.Helper()

	snapshotCount, err := store.client.RunSnapshot.Query().
		Where(entrunsnapshot.RunIDEQ(runID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, snapshotCount)

	journalCount, err := store.client.RunJournal.Query().
		Where(entrunjournal.RunIDEQ(runID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, journalCount)

	stepCount, err := store.client.RunStep.Query().
		Where(entrunstep.RunIDEQ(runID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, stepCount)
}

func assertResidualRunPresent(t *testing.T, ctx context.Context, store *EntStore, runID string) {
	t.Helper()

	snapshotCount, err := store.client.RunSnapshot.Query().
		Where(entrunsnapshot.RunIDEQ(runID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, snapshotCount)

	journalCount, err := store.client.RunJournal.Query().
		Where(entrunjournal.RunIDEQ(runID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, journalCount)

	stepCount, err := store.client.RunStep.Query().
		Where(entrunstep.RunIDEQ(runID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, stepCount)
}
