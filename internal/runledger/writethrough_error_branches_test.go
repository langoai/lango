package runledger

import (
	"context"
	"errors"
	"testing"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestWorkflowWriteThrough_CreateRun_StageShadowDelegatesWithoutLedgerSideEffects(t *testing.T) {
	ctx := context.Background()
	ledger := NewMemoryStore()
	projection := &recordingWorkflowProjectionStore{}
	wt := NewWorkflowWriteThrough(ledger, projection, RolloutConfig{Stage: StageShadow})

	runID, err := wt.CreateRun(ctx, &workflow.Workflow{Name: "wf-shadow", Description: "shadow only"})
	require.NoError(t, err)
	require.Equal(t, "projection-wf-shadow", runID)

	_, err = ledger.GetJournalEvents(ctx, runID)
	require.Error(t, err)

	runs, err := ledger.ListRuns(ctx, 10)
	require.NoError(t, err)
	require.Empty(t, runs)
	require.Nil(t, projection.runStatus)
	require.Empty(t, projection.runStatusUpdates)
}

func TestWorkflowWriteThrough_CreateRun_UpdateRunStatusFailureRecordsDegradedProjectionSync(t *testing.T) {
	ctx := context.Background()
	ledger := NewMemoryStore()
	projectionErr := errors.New("projection status write failed")
	projection := &failUpdateRunStatusProjectionStore{err: projectionErr}
	wt := NewWorkflowWriteThrough(ledger, projection, RolloutConfig{Stage: StageWriteThrough})

	runID, err := wt.CreateRun(ctx, &workflow.Workflow{
		Name: "wf-degraded-create",
		Steps: []workflow.Step{
			{ID: "step-1", Agent: "operator", Prompt: "do work"},
		},
	})
	require.ErrorIs(t, err, projectionErr)
	require.Empty(t, runID)
	require.NotEmpty(t, projection.runID)

	events, err := ledger.GetJournalEvents(ctx, projection.runID)
	require.NoError(t, err)
	require.Contains(t, eventTypes(events), EventRunCreated)
	require.Contains(t, eventTypes(events), EventPlanAttached)
	require.Equal(t, 1, countProjectionSyncStatus(t, events, "degraded"))
}

func TestWorkflowWriteThrough_UpdateStepStatus_FailedAndUnknownBranches(t *testing.T) {
	ctx := context.Background()
	ledger := NewMemoryStore()
	projection := &recordingWorkflowProjectionStore{}
	wt := NewWorkflowWriteThrough(ledger, projection, RolloutConfig{Stage: StageWriteThrough})

	runID, err := wt.CreateRun(ctx, &workflow.Workflow{
		Name: "wf-step-branches",
		Steps: []workflow.Step{
			{ID: "failed-step", Agent: "operator", Prompt: "fail"},
			{ID: "unknown-step", Agent: "operator", Prompt: "pause"},
		},
	})
	require.NoError(t, err)

	require.NoError(t, wt.UpdateStepStatus(ctx, runID, "failed-step", "failed", "", "failed reason"))
	require.NoError(t, wt.UpdateStepStatus(ctx, runID, "unknown-step", "paused", "", ""))

	snap, err := ledger.GetRunSnapshot(ctx, runID)
	require.NoError(t, err)
	require.Equal(t, StepStatusFailed, snap.FindStep("failed-step").Status)
	require.Equal(t, StepStatusPending, snap.FindStep("unknown-step").Status)
	require.Equal(t, "failed reason", snap.CurrentBlocker)

	events, err := ledger.GetJournalEvents(ctx, runID)
	require.NoError(t, err)
	require.Equal(t, 1, countEventType(events, EventStepValidationFailed))
	require.Equal(t, 0, countEventType(events, EventStepStarted))
	require.Equal(t, 0, countEventType(events, EventStepResultProposed))
	require.Contains(t, projection.stepStatusUpdates, runID+":failed-step:failed::failed reason")
	require.Contains(t, projection.stepStatusUpdates, runID+":unknown-step:paused::")
}

func TestDetectWorkflowProjectionDrift_MismatchCases(t *testing.T) {
	ctx := context.Background()

	t.Run("projection missing", func(t *testing.T) {
		ledger, runID := createWorkflowLedgerRun(t, ctx)
		drift, err := DetectWorkflowProjectionDrift(ctx, ledger, failingWorkflowProjectionStore{err: errors.New("missing")}, runID)
		require.NoError(t, err)
		require.Equal(t, &ProjectionDrift{
			RunID:  runID,
			Target: "workflow",
			Reason: "workflow projection missing",
		}, drift)
	})

	t.Run("step count mismatch", func(t *testing.T) {
		ledger, runID := createWorkflowLedgerRun(t, ctx)
		drift, err := DetectWorkflowProjectionDrift(ctx, ledger, workflowStatusProjectionStore{
			status: &workflow.RunStatus{RunID: runID, Status: string(RunStatusRunning)},
		}, runID)
		require.NoError(t, err)
		require.Equal(t, "step count mismatch: ledger=1 projection=0", drift.Reason)
	})

	t.Run("step status mismatch", func(t *testing.T) {
		ledger, runID := createWorkflowLedgerRun(t, ctx)
		drift, err := DetectWorkflowProjectionDrift(ctx, ledger, workflowStatusProjectionStore{
			status: &workflow.RunStatus{
				RunID:  runID,
				Status: string(RunStatusRunning),
				StepStatuses: []workflow.StepStatus{{
					StepID: "step-1",
					Status: "completed",
				}},
			},
		}, runID)
		require.NoError(t, err)
		require.Equal(t, "step step-1 mismatch: ledger=pending projection=completed", drift.Reason)
	})
}

func TestWorkflowWriteThrough_CompleteRun_ErrorBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("projection failure records degraded after terminal event", func(t *testing.T) {
		ledger := NewMemoryStore()
		projection := &recordingWorkflowProjectionStore{}
		wt := NewWorkflowWriteThrough(ledger, projection, RolloutConfig{Stage: StageWriteThrough})
		runID, err := wt.CreateRun(ctx, &workflow.Workflow{Name: "wf-complete-projection-fails"})
		require.NoError(t, err)

		projection.err = errors.New("complete projection failed")
		err = wt.CompleteRun(ctx, runID, "completed", "")
		require.ErrorIs(t, err, projection.err)

		events, eventsErr := ledger.GetJournalEvents(ctx, runID)
		require.NoError(t, eventsErr)
		require.Contains(t, eventTypes(events), EventRunCompleted)
		require.Equal(t, 1, countProjectionSyncStatus(t, events, "degraded"))
	})

	t.Run("prune failure stops before projection completion", func(t *testing.T) {
		ledger := &failingRunLedgerStore{MemoryStore: NewMemoryStore(), err: errors.New("prune failed")}
		projection := &recordingWorkflowProjectionStore{}
		wt := NewWorkflowWriteThrough(ledger, projection, RolloutConfig{Stage: StageWriteThrough}).WithMaxHistory(1)
		runID, err := wt.CreateRun(ctx, &workflow.Workflow{Name: "wf-prune-fails"})
		require.NoError(t, err)

		ledger.failPrune = true
		err = wt.CompleteRun(ctx, runID, "completed", "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "prune old runs")
		require.Empty(t, projection.completedRuns)
	})
}

func TestReplayWorkflowProjection_RecreatesMissingProjectionAndLedgerOnlyFailedStep(t *testing.T) {
	ctx := context.Background()
	ledger := NewMemoryStore()
	runID := "wf-replay-extra"
	require.NoError(t, ledger.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{
			Goal:             "replay workflow",
			OriginalRequest:  "replay missing projection",
			SourceKind:       "workflow",
			SourceDescriptor: marshalPayload(workflow.Workflow{Name: "wf-replay-extra"}),
		}),
	}))
	require.NoError(t, ledger.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{Steps: []Step{
			{StepID: "step-1", Goal: "do work", OwnerAgent: "operator", Status: StepStatusPending},
			{StepID: "step-extra", Goal: "extra work", OwnerAgent: "reviewer", Status: StepStatusPending},
		}}),
	}))
	require.NoError(t, ledger.RecordValidationResult(ctx, runID, "step-extra", ValidationResult{
		Passed: false,
		Reason: "blocked reason",
	}))

	projection := &missingRunStatusProjectionStore{err: errors.New("projection missing")}
	err := ReplayWorkflowProjection(ctx, ledger, projection, runID, &workflow.Workflow{
		Name: "wf-replay-extra",
		Steps: []workflow.Step{
			{ID: "step-1", Agent: "operator", Prompt: "do work"},
		},
	})
	require.NoError(t, err)

	require.NotNil(t, projection.runStatus)
	require.Equal(t, runID, projection.runStatus.RunID)
	require.Contains(t, projection.createdSteps, runID+":step-1:do work")
	require.Contains(t, projection.createdSteps, runID+":step-extra:extra work")
	require.Contains(t, projection.stepStatusUpdates, runID+":step-extra:failed::blocked reason")
	require.Contains(t, projection.runStatusUpdates, runID+":"+string(RunStatusRunning))
}

func TestBackgroundWriteThrough_PrepareTaskWithID_StageShadowNoLedgerSideEffects(t *testing.T) {
	ctx := context.Background()
	ledger := NewMemoryStore()
	projection := NewBackgroundWriteThrough(ledger, RolloutConfig{Stage: StageShadow})

	require.NoError(t, projection.PrepareTaskWithID(ctx, "background prompt", background.Origin{Session: "session-1"}, "bg-shadow"))

	_, err := ledger.GetJournalEvents(ctx, "bg-shadow")
	require.Error(t, err)

	runs, err := ledger.ListRuns(ctx, 10)
	require.NoError(t, err)
	require.Empty(t, runs)
}

func TestBackgroundWriteThrough_SyncTask_StageShadowNoLedgerSideEffects(t *testing.T) {
	ctx := context.Background()
	ledger := NewMemoryStore()
	projection := NewBackgroundWriteThrough(ledger, RolloutConfig{Stage: StageShadow})

	require.NoError(t, projection.SyncTask(ctx, background.TaskSnapshot{
		ID:         "bg-shadow-sync",
		StatusText: "done",
		Result:     "result body",
	}))

	_, err := ledger.GetJournalEvents(ctx, "bg-shadow-sync")
	require.Error(t, err)

	runs, err := ledger.ListRuns(ctx, 10)
	require.NoError(t, err)
	require.Empty(t, runs)
}

func TestBackgroundWriteThrough_PrepareTaskWithID_ErrorBranches(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name           string
		failEventType  JournalEventType
		failSnapshot   bool
		wantErr        string
		wantEventCount int
	}{
		{
			name:           "run created append fails",
			failEventType:  EventRunCreated,
			wantErr:        "append background run_created",
			wantEventCount: 0,
		},
		{
			name:           "plan attached append fails",
			failEventType:  EventPlanAttached,
			wantErr:        "append background plan_attached",
			wantEventCount: 1,
		},
		{
			name:           "snapshot materialization fails",
			failSnapshot:   true,
			wantErr:        "materialize background snapshot",
			wantEventCount: 2,
		},
		{
			name:           "projection sync append fails",
			failEventType:  EventProjectionSynced,
			wantErr:        "append background projection_synced",
			wantEventCount: 2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ledger := &failingRunLedgerStore{
				MemoryStore:   NewMemoryStore(),
				failEventType: tc.failEventType,
				failSnapshot:  tc.failSnapshot,
				err:           errors.New("ledger write failed"),
			}
			projection := NewBackgroundWriteThrough(ledger, RolloutConfig{Stage: StageWriteThrough})

			err := projection.PrepareTaskWithID(ctx, "background prompt", background.Origin{Session: "session-1"}, "bg-error")
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)

			events, eventsErr := ledger.MemoryStore.GetJournalEvents(ctx, "bg-error")
			if tc.wantEventCount == 0 {
				require.Error(t, eventsErr)
				return
			}
			require.NoError(t, eventsErr)
			require.Len(t, events, tc.wantEventCount)
		})
	}
}

func TestBackgroundWriteThrough_SyncTask_ErrorBranches(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name          string
		statusText    string
		configureFail func(*failingRunLedgerStore)
		wantErr       string
	}{
		{
			name:       "running append fails",
			statusText: "running",
			configureFail: func(store *failingRunLedgerStore) {
				store.failEventType = EventStepStarted
			},
			wantErr: "ledger write failed",
		},
		{
			name:       "done validation fails",
			statusText: "done",
			configureFail: func(store *failingRunLedgerStore) {
				store.failRecordValidation = true
			},
			wantErr: "ledger write failed",
		},
		{
			name:       "failed prune fails",
			statusText: "failed",
			configureFail: func(store *failingRunLedgerStore) {
				store.failPrune = true
			},
			wantErr: "ledger write failed",
		},
		{
			name:       "final projection sync append fails",
			statusText: "paused",
			configureFail: func(store *failingRunLedgerStore) {
				store.failEventType = EventProjectionSynced
			},
			wantErr: "ledger write failed",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ledger := &failingRunLedgerStore{MemoryStore: NewMemoryStore(), err: errors.New("ledger write failed")}
			projection := NewBackgroundWriteThrough(ledger, RolloutConfig{Stage: StageWriteThrough}).WithMaxHistory(1)
			require.NoError(t, projection.PrepareTaskWithID(ctx, "background prompt", background.Origin{Session: "session-1"}, "bg-sync-error"))
			beforeEvents, err := ledger.MemoryStore.GetJournalEvents(ctx, "bg-sync-error")
			require.NoError(t, err)
			beforeSynced := countProjectionSyncStatus(t, beforeEvents, "synced")

			tc.configureFail(ledger)
			err = projection.SyncTask(ctx, background.TaskSnapshot{
				ID:         "bg-sync-error",
				StatusText: tc.statusText,
				Result:     "result body",
				Error:      "error body",
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)

			afterEvents, eventsErr := ledger.MemoryStore.GetJournalEvents(ctx, "bg-sync-error")
			require.NoError(t, eventsErr)
			require.Equal(t, beforeSynced, countProjectionSyncStatus(t, afterEvents, "synced"))
		})
	}
}

func createWorkflowLedgerRun(t *testing.T, ctx context.Context) (*MemoryStore, string) {
	t.Helper()

	ledger := NewMemoryStore()
	wt := NewWorkflowWriteThrough(ledger, &recordingWorkflowProjectionStore{}, RolloutConfig{Stage: StageWriteThrough})
	runID, err := wt.CreateRun(ctx, &workflow.Workflow{
		Name: "wf-drift",
		Steps: []workflow.Step{
			{ID: "step-1", Agent: "operator", Prompt: "do work"},
		},
	})
	require.NoError(t, err)
	return ledger, runID
}

func countEventType(events []JournalEvent, eventType JournalEventType) int {
	count := 0
	for _, event := range events {
		if event.Type == eventType {
			count++
		}
	}
	return count
}

type failUpdateRunStatusProjectionStore struct {
	recordingWorkflowProjectionStore
	err   error
	runID string
}

func (f *failUpdateRunStatusProjectionStore) CreateRunWithID(_ context.Context, runID string, _ *workflow.Workflow) error {
	f.runID = runID
	return nil
}

func (f *failUpdateRunStatusProjectionStore) UpdateRunStatus(_ context.Context, _ string, _ string) error {
	return f.err
}

type workflowStatusProjectionStore struct {
	failingWorkflowProjectionStore
	status *workflow.RunStatus
}

func (w workflowStatusProjectionStore) GetRunStatus(_ context.Context, _ string) (*workflow.RunStatus, error) {
	return w.status, nil
}

type missingRunStatusProjectionStore struct {
	recordingWorkflowProjectionStore
	err error
}

func (m *missingRunStatusProjectionStore) GetRunStatus(_ context.Context, _ string) (*workflow.RunStatus, error) {
	return nil, m.err
}

type failingRunLedgerStore struct {
	*MemoryStore
	failEventType        JournalEventType
	failSnapshot         bool
	failRecordValidation bool
	failPrune            bool
	err                  error
}

func (f *failingRunLedgerStore) AppendJournalEvent(ctx context.Context, event JournalEvent) error {
	if event.Type == f.failEventType {
		return f.err
	}
	return f.MemoryStore.AppendJournalEvent(ctx, event)
}

func (f *failingRunLedgerStore) GetRunSnapshot(ctx context.Context, runID string) (*RunSnapshot, error) {
	if f.failSnapshot {
		return nil, f.err
	}
	return f.MemoryStore.GetRunSnapshot(ctx, runID)
}

func (f *failingRunLedgerStore) RecordValidationResult(ctx context.Context, runID, stepID string, result ValidationResult) error {
	if f.failRecordValidation {
		return f.err
	}
	return f.MemoryStore.RecordValidationResult(ctx, runID, stepID, result)
}

func (f *failingRunLedgerStore) PruneOldRuns(ctx context.Context, maxKeep int) error {
	if f.failPrune {
		return f.err
	}
	return f.MemoryStore.PruneOldRuns(ctx, maxKeep)
}
