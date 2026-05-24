package runledger

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/enttest"
	entrunsnapshot "github.com/langoai/lango/internal/ent/runsnapshot"
	entrunstep "github.com/langoai/lango/internal/ent/runstep"
)

func newEntStoreGetRunSnapshotAppliesTailAndRewritesStepProjectionEntStore(t *testing.T) (*EntStore, context.Context) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?_fk=1", filepath.Join(t.TempDir(), "runledger-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8.db"))
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	return NewEntStore(client), context.Background()
}

func entStoreGetRunSnapshotAppliesTailAndRewritesStepProjectionAppendCreatedAndPlan(t *testing.T, ctx context.Context, store *EntStore, runID, sessionKey string) {
	t.Helper()

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{
			SessionKey:      sessionKey,
			OriginalRequest: "cover ent store branches",
			Goal:            "exercise run ledger storage",
		}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{
				{
					StepID:     "step-1",
					Index:      0,
					Goal:       "collect evidence",
					OwnerAgent: "tester",
					Status:     StepStatusPending,
					Validator:  ValidatorSpec{Type: ValidatorArtifactExists, Target: "artifact.txt"},
					MaxRetries: DefaultMaxRetries,
				},
				{
					StepID:     "step-2",
					Index:      1,
					Goal:       "verify evidence",
					OwnerAgent: "reviewer",
					Status:     StepStatusPending,
					Validator:  ValidatorSpec{Type: ValidatorTestPass, Target: "./internal/runledger"},
					MaxRetries: DefaultMaxRetries,
				},
			},
			AcceptanceCriteria: []AcceptanceCriterion{{
				Description: "storage branches are covered",
				Validator:   ValidatorSpec{Type: ValidatorTestPass, Target: "./internal/runledger"},
			}},
		}),
	}))
}

func TestEntStoreGetRunSnapshotAppliesTailAndRewritesStepProjection(t *testing.T) {
	store, ctx := newEntStoreGetRunSnapshotAppliesTailAndRewritesStepProjectionEntStore(t)
	entStoreGetRunSnapshotAppliesTailAndRewritesStepProjectionAppendCreatedAndPlan(t, ctx, store, "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-tail", "session-tail")

	initial, err := store.GetRunSnapshot(ctx, "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-tail")
	require.NoError(t, err)
	require.Equal(t, int64(2), initial.LastJournalSeq)
	require.Len(t, initial.Steps, 2)

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-tail",
		Type:    EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "step-1", OwnerAgent: "tester"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-tail",
		Type:  EventStepResultProposed,
		Payload: marshalPayload(StepResultProposedPayload{
			StepID: "step-1",
			Result: "artifact produced",
			Evidence: []Evidence{{
				Type:    "file",
				Content: "artifact.txt",
			}},
		}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-tail",
		Type:    EventStepValidationPassed,
		Payload: marshalPayload(StepValidationPassedPayload{StepID: "step-1", Result: ValidationResult{Passed: true, Reason: "ok"}}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-tail",
		Type:    EventCriterionMet,
		Payload: marshalPayload(CriterionMetPayload{Index: 0, Description: "storage branches are covered"}),
	}))

	updated, err := store.GetRunSnapshot(ctx, "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-tail")
	require.NoError(t, err)
	require.Equal(t, int64(6), updated.LastJournalSeq)
	require.Len(t, updated.Steps, 2)
	assert.Equal(t, StepStatusCompleted, updated.Steps[0].Status)
	assert.Equal(t, "artifact produced", updated.Steps[0].Result)
	require.Len(t, updated.Steps[0].Evidence, 1)
	assert.Equal(t, "artifact.txt", updated.Steps[0].Evidence[0].Content)
	require.Len(t, updated.AcceptanceState, 1)
	assert.True(t, updated.AcceptanceState[0].Met)

	rows, err := store.client.RunStep.Query().
		Where(entrunstep.RunIDEQ("run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-tail")).
		Order(entrunstep.ByStepIndex()).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, entrunstep.StatusCompleted, rows[0].Status)
	assert.Equal(t, "artifact produced", rows[0].Result)
	assert.Equal(t, entrunstep.StatusPending, rows[1].Status)

	again, err := store.GetRunSnapshot(ctx, "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-tail")
	require.NoError(t, err)
	assert.Equal(t, updated.LastJournalSeq, again.LastJournalSeq)

	count, err := store.client.RunStep.Query().
		Where(entrunstep.RunIDEQ("run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-tail")).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestEntStoreMaterializeAndSnapshotReportMalformedPayloads(t *testing.T) {
	store, ctx := newEntStoreGetRunSnapshotAppliesTailAndRewritesStepProjectionEntStore(t)

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-bad-journal",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "session-bad", Goal: "bad journal"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-bad-journal",
		Type:    EventPlanAttached,
		Payload: json.RawMessage(`{`),
	}))

	events, err := store.GetJournalEvents(ctx, "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-bad-journal")
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, json.RawMessage(`{`), events[1].Payload)

	snapshot, err := store.MaterializeRunSnapshot(ctx, "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-bad-journal")
	require.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), "apply event seq 2")
	assert.Contains(t, err.Error(), "unmarshal plan_attached")

	snapshot, err = store.GetRunSnapshot(ctx, "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-bad-journal")
	require.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), "unmarshal plan_attached")
}

func TestEntStoreListsFilterLimitEmptyAndMaxSeq(t *testing.T) {
	store, ctx := newEntStoreGetRunSnapshotAppliesTailAndRewritesStepProjectionEntStore(t)

	snapshots := []*RunSnapshot{
		{
			RunID:          "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-session-a-1",
			SessionKey:     "session-a",
			Goal:           "first session a run",
			Status:         RunStatusRunning,
			LastJournalSeq: 4,
			Steps: []Step{{
				StepID:     "a1-step",
				Goal:       "a1",
				Status:     StepStatusCompleted,
				Validator:  ValidatorSpec{Type: ValidatorBuildPass},
				MaxRetries: DefaultMaxRetries,
			}},
		},
		{
			RunID:          "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-session-a-2",
			SessionKey:     "session-a",
			Goal:           "second session a run",
			Status:         RunStatusCompleted,
			LastJournalSeq: 9,
			Steps: []Step{{
				StepID:     "a2-step",
				Goal:       "a2",
				Status:     StepStatusCompleted,
				Validator:  ValidatorSpec{Type: ValidatorTestPass},
				MaxRetries: DefaultMaxRetries,
			}},
		},
		{
			RunID:          "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-session-b-1",
			SessionKey:     "session-b",
			Goal:           "session b run",
			Status:         RunStatusFailed,
			LastJournalSeq: 6,
		},
	}
	for _, snapshot := range snapshots {
		require.NoError(t, store.UpdateCachedSnapshot(ctx, snapshot))
	}

	allSessionA, err := store.ListRunSummariesBySession(ctx, "session-a", 0)
	require.NoError(t, err)
	require.Len(t, allSessionA, 2)
	assert.ElementsMatch(t, []string{"run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-session-a-1", "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-session-a-2"}, []string{allSessionA[0].RunID, allSessionA[1].RunID})
	for _, summary := range allSessionA {
		assert.NotEqual(t, "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-session-b-1", summary.RunID)
	}

	limitedSessionA, err := store.ListRunSummariesBySession(ctx, "session-a", 1)
	require.NoError(t, err)
	require.Len(t, limitedSessionA, 1)
	assert.Contains(t, []string{"run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-session-a-1", "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-session-a-2"}, limitedSessionA[0].RunID)

	empty, err := store.ListRunSummariesBySession(ctx, "missing-session", 5)
	require.NoError(t, err)
	assert.Empty(t, empty)

	maxSeq, err := store.MaxJournalSeqForSession(ctx, "session-a")
	require.NoError(t, err)
	assert.Equal(t, int64(9), maxSeq)

	maxSeq, err = store.MaxJournalSeqForSession(ctx, "missing-session")
	require.NoError(t, err)
	assert.Zero(t, maxSeq)

	allRuns, err := store.ListRuns(ctx, 2)
	require.NoError(t, err)
	require.Len(t, allRuns, 2)
}

func TestEntStoreMalformedPersistedSnapshotsFailListPaths(t *testing.T) {
	store, ctx := newEntStoreGetRunSnapshotAppliesTailAndRewritesStepProjectionEntStore(t)

	_, err := store.client.RunSnapshot.Create().
		SetRunID("run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-list-corrupt").
		SetSessionKey("session-corrupt").
		SetStatus(entrunsnapshot.StatusRunning).
		SetGoal("corrupt list row").
		SetSnapshotData(`{`).
		SetLastJournalSeq(1).
		Save(ctx)
	require.NoError(t, err)

	runs, err := store.ListRuns(ctx, 10)
	require.Error(t, err)
	assert.Nil(t, runs)
	assert.Contains(t, err.Error(), "listed snapshot run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-list-corrupt")

	sessionRuns, err := store.ListRunSummariesBySession(ctx, "session-corrupt", 10)
	require.Error(t, err)
	assert.Nil(t, sessionRuns)
	assert.Contains(t, err.Error(), "session snapshot run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-list-corrupt")
}

func TestEntStorePruneNoopBranchesPreserveRows(t *testing.T) {
	store, ctx := newEntStoreGetRunSnapshotAppliesTailAndRewritesStepProjectionEntStore(t)

	for _, runID := range []string{"run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-prune-1", "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-prune-2"} {
		entStoreGetRunSnapshotAppliesTailAndRewritesStepProjectionAppendCreatedAndPlan(t, ctx, store, runID, "session-prune")
		snapshot, err := store.GetRunSnapshot(ctx, runID)
		require.NoError(t, err)
		require.NoError(t, store.UpdateCachedSnapshot(ctx, snapshot))
	}

	require.NoError(t, store.PruneOldRuns(ctx, 0))
	require.NoError(t, store.PruneOldRuns(ctx, -1))
	require.NoError(t, store.PruneOldRuns(ctx, 3))
	require.NoError(t, store.PruneOldRuns(ctx, 1))

	runs, err := store.ListRuns(ctx, 10)
	require.NoError(t, err)
	require.Len(t, runs, 2)
	assert.ElementsMatch(t, []string{"run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-prune-1", "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-prune-2"}, []string{runs[0].RunID, runs[1].RunID})

	journalEvents, err := store.GetJournalEvents(ctx, "run-knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-prune-1")
	require.NoError(t, err)
	assert.Len(t, journalEvents, 2)
}
