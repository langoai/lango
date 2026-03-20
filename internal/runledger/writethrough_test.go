package runledger

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
