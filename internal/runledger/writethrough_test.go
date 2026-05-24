package runledger

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/workflow"
	"go.uber.org/zap"
)

func TestWorkflowWriteThrough_CreateRun_UsesCanonicalRunID(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	ledger := NewEntStore(client)
	projection := workflow.NewStateStore(client, zap.NewNop().Sugar())
	wt := NewWorkflowWriteThrough(ledger, projection, RolloutConfig{
		Stage: StageWriteThrough,
	})

	runID, err := wt.CreateRun(context.Background(), &workflow.Workflow{
		Name:        "wf-1",
		Description: "test workflow",
		Steps: []workflow.Step{
			{ID: "step-1", Agent: "operator", Prompt: "do work"},
		},
	})
	require.NoError(t, err)

	snap, err := ledger.GetRunSnapshot(context.Background(), runID)
	require.NoError(t, err)
	assert.Equal(t, runID, snap.RunID)

	status, err := projection.GetRunStatus(context.Background(), runID)
	require.NoError(t, err)
	assert.Equal(t, runID, status.RunID)
	assert.Equal(t, "wf-1", status.WorkflowName)
	assert.Equal(t, "running", status.Status)
}

func TestWorkflowWriteThrough_CreateRun_RecordsDegradedProjectionOnFailure(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	ledger := NewEntStore(client)
	wt := NewWorkflowWriteThrough(ledger, failingWorkflowProjectionStore{
		err: errors.New("projection create failed"),
	}, RolloutConfig{Stage: StageWriteThrough})

	_, err := wt.CreateRun(context.Background(), &workflow.Workflow{
		Name:        "wf-fail",
		Description: "broken projection",
	})
	require.Error(t, err)

	runs, listErr := ledger.ListRuns(context.Background(), 10)
	require.NoError(t, listErr)
	require.Len(t, runs, 1)

	events, eventsErr := ledger.GetJournalEvents(context.Background(), runs[0].RunID)
	require.NoError(t, eventsErr)

	foundDegraded := false
	for _, event := range events {
		if event.Type != EventProjectionSynced {
			continue
		}
		var payload ProjectionSyncPayload
		require.NoError(t, json.Unmarshal(event.Payload, &payload))
		if payload.Status == "degraded" {
			foundDegraded = true
			assert.Equal(t, "workflow", payload.Target)
			assert.Contains(t, payload.Error, "projection create failed")
		}
	}
	assert.True(t, foundDegraded)
}

func TestDetectAndReplayWorkflowProjectionDrift(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	ledger := NewEntStore(client)
	projection := workflow.NewStateStore(client, zap.NewNop().Sugar())
	wt := NewWorkflowWriteThrough(ledger, projection, RolloutConfig{Stage: StageWriteThrough})
	ctx := context.Background()

	wf := &workflow.Workflow{
		Name:        "wf-replay",
		Description: "rebuild projection",
		Steps: []workflow.Step{
			{ID: "step-1", Agent: "operator", Prompt: "do work"},
		},
	}

	runID, err := wt.CreateRun(ctx, wf)
	require.NoError(t, err)

	// Corrupt the projection status directly.
	require.NoError(t, projection.UpdateRunStatus(ctx, runID, "failed"))

	drift, err := DetectWorkflowProjectionDrift(ctx, ledger, projection, runID)
	require.NoError(t, err)
	require.NotNil(t, drift)
	assert.Contains(t, drift.Reason, "status mismatch")

	require.NoError(t, ReplayWorkflowProjection(ctx, ledger, projection, runID, wf))

	drift, err = DetectWorkflowProjectionDrift(ctx, ledger, projection, runID)
	require.NoError(t, err)
	assert.Nil(t, drift)
}

func TestBackgroundWriteThrough_SnapshotMatchesBackgroundTaskStatus(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	ledger := NewEntStore(client)
	projection := NewBackgroundWriteThrough(ledger, RolloutConfig{
		Stage: StageWriteThrough,
	})
	mgr := background.NewManager(&backgroundTestRunner{result: "done"}, nil, 5, time.Minute, zap.NewNop().Sugar()).
		WithProjection(projection)

	runID, err := mgr.Submit(context.Background(), "background prompt", background.Origin{
		Session: "session-1",
	})
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)

	snap, err := ledger.GetRunSnapshot(context.Background(), runID)
	require.NoError(t, err)
	assert.Equal(t, RunStatusCompleted, snap.Status)
	require.Len(t, snap.Steps, 1)
	assert.Equal(t, StepStatusCompleted, snap.Steps[0].Status)
}

func TestBackgroundWriteThrough_PrepareTaskWithID_UsesProvidedID(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	ledger := NewEntStore(client)
	projection := NewBackgroundWriteThrough(ledger, RolloutConfig{
		Stage: StageWriteThrough,
	})

	err := projection.PrepareTaskWithID(context.Background(), "background prompt", background.Origin{
		Session: "session-1",
	}, "run-fixed")
	require.NoError(t, err)

	snap, err := ledger.GetRunSnapshot(context.Background(), "run-fixed")
	require.NoError(t, err)
	assert.Equal(t, "run-fixed", snap.RunID)
	assert.Equal(t, RunStatusRunning, snap.Status)
}

func TestRolloutConfig_ReadModePredicates(t *testing.T) {
	t.Parallel()

	assert.True(t, RolloutConfig{Stage: StageShadow}.IsShadow())
	assert.False(t, RolloutConfig{Stage: StageWriteThrough}.IsShadow())
	assert.False(t, RolloutConfig{Stage: StageWriteThrough}.IsAuthoritativeRead())
	assert.True(t, RolloutConfig{Stage: StageAuthoritativeRead}.IsAuthoritativeRead())
	assert.True(t, RolloutConfig{Stage: StageProjectionRetired}.IsAuthoritativeRead())
}

func TestWorkflowWriteThrough_DelegatesStatusAndReadMethods(t *testing.T) {
	ctx := context.Background()
	ledger := NewMemoryStore()
	projection := &recordingWorkflowProjectionStore{
		runStatus: &workflow.RunStatus{RunID: "run-1", WorkflowName: "wf", Status: "running"},
		stepResults: map[string]string{
			"step-1": "done",
		},
		runs: []workflow.RunStatus{{RunID: "run-1", WorkflowName: "wf", Status: "running"}},
	}
	wt := NewWorkflowWriteThrough(ledger, projection, RolloutConfig{Stage: StageWriteThrough}).WithMaxHistory(3)
	require.Same(t, wt, wt.WithMaxHistory(3))
	require.Equal(t, 3, wt.maxKeep)

	require.NoError(t, wt.UpdateRunStatus(ctx, "run-1", "running"))
	require.NoError(t, wt.CreateStepRun(ctx, "run-1", workflow.Step{ID: "step-1", Agent: "operator"}, "rendered"))

	gotStatus, err := wt.GetRunStatus(ctx, "run-1")
	require.NoError(t, err)
	require.Equal(t, projection.runStatus, gotStatus)

	gotResults, err := wt.GetStepResults(ctx, "run-1")
	require.NoError(t, err)
	require.Equal(t, projection.stepResults, gotResults)

	gotRuns, err := wt.ListRuns(ctx, 5)
	require.NoError(t, err)
	require.Equal(t, projection.runs, gotRuns)
	require.Equal(t, []string{"run-1:running"}, projection.runStatusUpdates)
	require.Equal(t, []string{"run-1:step-1:rendered"}, projection.createdSteps)
}

func TestWorkflowWriteThrough_UpdateStepStatus_MirrorsLedgerEvents(t *testing.T) {
	ctx := context.Background()
	ledger := NewMemoryStore()
	projection := &recordingWorkflowProjectionStore{}
	wt := NewWorkflowWriteThrough(ledger, projection, RolloutConfig{Stage: StageWriteThrough})

	runID, err := wt.CreateRun(ctx, &workflow.Workflow{
		Name:        "wf-steps",
		Description: "step transitions",
		Steps: []workflow.Step{
			{ID: "step-1", Agent: "operator", Prompt: "do work"},
		},
	})
	require.NoError(t, err)

	require.NoError(t, wt.UpdateStepStatus(ctx, runID, "step-1", "running", "", ""))
	require.NoError(t, wt.UpdateStepStatus(ctx, runID, "step-1", "completed", "result body", ""))

	snap, err := ledger.GetRunSnapshot(ctx, runID)
	require.NoError(t, err)
	require.Len(t, snap.Steps, 1)
	require.Equal(t, StepStatusCompleted, snap.Steps[0].Status)
	require.Equal(t, "result body", snap.Steps[0].Result)
	require.Equal(t, []string{
		runID + ":step-1:running::",
		runID + ":step-1:completed:result body:",
	}, projection.stepStatusUpdates)
}

func TestWorkflowWriteThrough_CompleteRun_AppendsTerminalEvents(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name      string
		status    string
		errMsg    string
		wantEvent JournalEventType
		wantRun   RunStatus
	}{
		{name: "completed", status: "completed", wantEvent: EventRunCompleted, wantRun: RunStatusCompleted},
		{name: "failed", status: "failed", errMsg: "boom", wantEvent: EventRunFailed, wantRun: RunStatusFailed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ledger := NewMemoryStore()
			projection := &recordingWorkflowProjectionStore{}
			wt := NewWorkflowWriteThrough(ledger, projection, RolloutConfig{Stage: StageWriteThrough})
			runID, err := wt.CreateRun(ctx, &workflow.Workflow{Name: "wf-" + tc.name})
			require.NoError(t, err)

			require.NoError(t, wt.CompleteRun(ctx, runID, tc.status, tc.errMsg))

			events, err := ledger.GetJournalEvents(ctx, runID)
			require.NoError(t, err)
			require.Contains(t, eventTypes(events), tc.wantEvent)
			snap, err := ledger.GetRunSnapshot(ctx, runID)
			require.NoError(t, err)
			require.Equal(t, tc.wantRun, snap.Status)
			require.Equal(t, []string{runID + ":" + tc.status + ":" + tc.errMsg}, projection.completedRuns)
		})
	}
}

func TestWorkflowWriteThrough_UpdateMethods_RecordDegradedProjectionOnFailure(t *testing.T) {
	ctx := context.Background()
	ledger := NewMemoryStore()
	projection := &recordingWorkflowProjectionStore{}
	wt := NewWorkflowWriteThrough(ledger, projection, RolloutConfig{Stage: StageWriteThrough})
	runID, err := wt.CreateRun(ctx, &workflow.Workflow{
		Name: "wf-failure",
		Steps: []workflow.Step{
			{ID: "step-1", Agent: "operator", Prompt: "do work"},
		},
	})
	require.NoError(t, err)

	projection.err = errors.New("projection unavailable")
	require.ErrorIs(t, wt.UpdateRunStatus(ctx, runID, "running"), projection.err)
	require.ErrorIs(t, wt.CreateStepRun(ctx, runID, workflow.Step{ID: "step-1"}, "rendered"), projection.err)
	require.ErrorIs(t, wt.UpdateStepStatus(ctx, runID, "step-1", "running", "", ""), projection.err)

	events, err := ledger.GetJournalEvents(ctx, runID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, countProjectionSyncStatus(t, events, "degraded"), 3)
}

func TestMapRunStepStatus_AllKnownStatuses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status StepStatus
		want   string
	}{
		{StepStatusCompleted, "completed"},
		{StepStatusFailed, "failed"},
		{StepStatusInProgress, "running"},
		{StepStatusInterrupted, "skipped"},
		{StepStatusVerifyPending, "running"},
		{StepStatusPending, "pending"},
		{StepStatus("CUSTOM_STATE"), "custom_state"},
	}
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, mapRunStepStatus(tt.status))
		})
	}
}

func TestBackgroundWriteThrough_WithMaxHistoryAndSyncTaskBranches(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name       string
		statusText string
		result     string
		errText    string
		wantRun    RunStatus
		wantStep   StepStatus
	}{
		{name: "pending", statusText: "pending", wantRun: RunStatusRunning, wantStep: StepStatusPending},
		{name: "running", statusText: "running", wantRun: RunStatusRunning, wantStep: StepStatusInProgress},
		{name: "done", statusText: "done", result: "done result", wantRun: RunStatusCompleted, wantStep: StepStatusCompleted},
		{name: "failed", statusText: "failed", errText: "runner failed", wantRun: RunStatusFailed, wantStep: StepStatusFailed},
		{name: "cancelled", statusText: "cancelled", wantRun: RunStatusFailed, wantStep: StepStatusPending},
		{name: "unknown", statusText: "paused", wantRun: RunStatusRunning, wantStep: StepStatusPending},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ledger := NewMemoryStore()
			projection := NewBackgroundWriteThrough(ledger, RolloutConfig{Stage: StageWriteThrough}).WithMaxHistory(2)
			require.Same(t, projection, projection.WithMaxHistory(2))
			require.Equal(t, 2, projection.maxKeep)

			runID := "bg-" + tc.name
			require.NoError(t, projection.PrepareTaskWithID(ctx, "background prompt", background.Origin{Session: "session-1"}, runID))
			require.NoError(t, projection.SyncTask(ctx, background.TaskSnapshot{
				ID:         runID,
				StatusText: tc.statusText,
				Result:     tc.result,
				Error:      tc.errText,
			}))

			snap, err := ledger.GetRunSnapshot(ctx, runID)
			require.NoError(t, err)
			require.Equal(t, tc.wantRun, snap.Status)
			require.Len(t, snap.Steps, 1)
			require.Equal(t, tc.wantStep, snap.Steps[0].Status)
		})
	}
}

type failingWorkflowProjectionStore struct {
	err error
}

func (f failingWorkflowProjectionStore) CreateRun(_ context.Context, _ *workflow.Workflow) (string, error) {
	return "", f.err
}

func (f failingWorkflowProjectionStore) CreateRunWithID(_ context.Context, _ string, _ *workflow.Workflow) error {
	return f.err
}

func (f failingWorkflowProjectionStore) UpdateRunStatus(_ context.Context, _ string, _ string) error {
	return f.err
}

func (f failingWorkflowProjectionStore) CompleteRun(_ context.Context, _ string, _ string, _ string) error {
	return f.err
}

func (f failingWorkflowProjectionStore) CreateStepRun(_ context.Context, _ string, _ workflow.Step, _ string) error {
	return f.err
}

func (f failingWorkflowProjectionStore) UpdateStepStatus(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ string,
	_ string,
) error {
	return f.err
}

func (f failingWorkflowProjectionStore) GetRunStatus(_ context.Context, _ string) (*workflow.RunStatus, error) {
	return nil, f.err
}

func (f failingWorkflowProjectionStore) GetStepResults(_ context.Context, _ string) (map[string]string, error) {
	return nil, f.err
}

func (f failingWorkflowProjectionStore) ListRuns(_ context.Context, _ int) ([]workflow.RunStatus, error) {
	return nil, f.err
}

type backgroundTestRunner struct {
	result string
	err    error
	delay  time.Duration
}

func (m *backgroundTestRunner) Run(_ context.Context, _ string, _ string) (string, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return m.result, m.err
}

type recordingWorkflowProjectionStore struct {
	err               error
	runStatus         *workflow.RunStatus
	stepResults       map[string]string
	runs              []workflow.RunStatus
	runStatusUpdates  []string
	createdSteps      []string
	stepStatusUpdates []string
	completedRuns     []string
}

func (r *recordingWorkflowProjectionStore) CreateRun(_ context.Context, wf *workflow.Workflow) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return "projection-" + wf.Name, nil
}

func (r *recordingWorkflowProjectionStore) CreateRunWithID(_ context.Context, runID string, _ *workflow.Workflow) error {
	if r.err != nil {
		return r.err
	}
	r.runStatus = &workflow.RunStatus{RunID: runID, Status: "pending"}
	return nil
}

func (r *recordingWorkflowProjectionStore) UpdateRunStatus(_ context.Context, runID string, status string) error {
	if r.err != nil {
		return r.err
	}
	r.runStatusUpdates = append(r.runStatusUpdates, runID+":"+status)
	return nil
}

func (r *recordingWorkflowProjectionStore) CompleteRun(_ context.Context, runID string, status string, errMsg string) error {
	if r.err != nil {
		return r.err
	}
	r.completedRuns = append(r.completedRuns, runID+":"+status+":"+errMsg)
	return nil
}

func (r *recordingWorkflowProjectionStore) CreateStepRun(_ context.Context, runID string, step workflow.Step, renderedPrompt string) error {
	if r.err != nil {
		return r.err
	}
	r.createdSteps = append(r.createdSteps, runID+":"+step.ID+":"+renderedPrompt)
	return nil
}

func (r *recordingWorkflowProjectionStore) UpdateStepStatus(
	_ context.Context,
	runID string,
	stepID string,
	status string,
	result string,
	errMsg string,
) error {
	if r.err != nil {
		return r.err
	}
	r.stepStatusUpdates = append(r.stepStatusUpdates, runID+":"+stepID+":"+status+":"+result+":"+errMsg)
	return nil
}

func (r *recordingWorkflowProjectionStore) GetRunStatus(_ context.Context, _ string) (*workflow.RunStatus, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.runStatus, nil
}

func (r *recordingWorkflowProjectionStore) GetStepResults(_ context.Context, _ string) (map[string]string, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.stepResults, nil
}

func (r *recordingWorkflowProjectionStore) ListRuns(_ context.Context, _ int) ([]workflow.RunStatus, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.runs, nil
}

func eventTypes(events []JournalEvent) []JournalEventType {
	types := make([]JournalEventType, 0, len(events))
	for _, event := range events {
		types = append(types, event.Type)
	}
	return types
}

func countProjectionSyncStatus(t *testing.T, events []JournalEvent, status string) int {
	t.Helper()

	count := 0
	for _, event := range events {
		if event.Type != EventProjectionSynced {
			continue
		}
		var payload ProjectionSyncPayload
		require.NoError(t, json.Unmarshal(event.Payload, &payload))
		if payload.Status == status {
			count++
		}
	}
	return count
}
