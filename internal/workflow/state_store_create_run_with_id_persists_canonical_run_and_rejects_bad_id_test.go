package workflow

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/ent/workflowrun"
	"github.com/langoai/lango/internal/ent/workflowsteprun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

func newStateStoreCreateRunWithIdPersistsCanonicalRunAndRejectsBadIdStateStore(t *testing.T) (*StateStore, *ent.Client) {
	t.Helper()
	name := strings.NewReplacer("/", "-", " ", "-").Replace(t.Name())
	client := enttest.Open(t, "sqlite3", "file:"+name+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	return NewStateStore(client, zap.NewNop().Sugar()), client
}

func TestStateStoreCreateRunWithIDPersistsCanonicalRunAndRejectsBadID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, client := newStateStoreCreateRunWithIdPersistsCanonicalRunAndRejectsBadIdStateStore(t)
	workflow := &Workflow{
		Name:        "newDeadLetterStatusLoaderWrapsBuildAppError0",
		Description: "state persistence",
		Steps: []Step{
			{ID: "first", Prompt: "collect"},
			{ID: "second", Prompt: "summarize"},
		},
	}

	err := store.CreateRunWithID(ctx, "not-a-uuid", workflow)
	require.Error(t, err)
	assert.ErrorContains(t, err, `parse run ID "not-a-uuid"`)

	runUUID := uuid.New()
	startedAfter := time.Now().Add(-time.Second)
	require.NoError(t, store.CreateRunWithID(ctx, runUUID.String(), workflow))

	run, err := client.WorkflowRun.Get(ctx, runUUID)
	require.NoError(t, err)
	assert.Equal(t, runUUID, run.ID)
	assert.Equal(t, "newDeadLetterStatusLoaderWrapsBuildAppError0", run.WorkflowName)
	assert.Equal(t, "state persistence", run.Description)
	assert.Equal(t, workflowrun.StatusPending, run.Status)
	assert.Equal(t, 2, run.TotalSteps)
	assert.Equal(t, 0, run.CompletedSteps)
	assert.False(t, run.StartedAt.Before(startedAfter))
	assert.Nil(t, run.CompletedAt)

	err = store.CreateRunWithID(ctx, runUUID.String(), workflow)
	require.Error(t, err)
	assert.ErrorContains(t, err, "create workflow run with id")
}

func TestStateStoreStatusAndResultsReflectPersistedStepTransitions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, client := newStateStoreCreateRunWithIdPersistsCanonicalRunAndRejectsBadIdStateStore(t)
	runID := uuid.New().String()
	workflow := &Workflow{
		Name:        "persisted-status",
		Description: "tracks step outcomes",
		Steps: []Step{
			{ID: "done", Agent: "operator", Prompt: "finish"},
			{ID: "failed", Agent: "executor", Prompt: "fail"},
			{ID: "skipped", Prompt: "skip"},
		},
	}
	require.NoError(t, store.CreateRunWithID(ctx, runID, workflow))
	require.NoError(t, store.UpdateRunStatus(ctx, runID, "running"))
	for _, step := range workflow.Steps {
		require.NoError(t, store.CreateStepRun(ctx, runID, step, "rendered "+step.ID))
	}

	require.NoError(t, store.UpdateStepStatus(ctx, runID, "done", "running", "", ""))
	require.NoError(t, store.UpdateStepStatus(ctx, runID, "done", "completed", "done-result", ""))
	require.NoError(t, store.UpdateStepStatus(ctx, runID, "failed", "running", "", ""))
	require.NoError(t, store.UpdateStepStatus(ctx, runID, "failed", "failed", "", "agent exploded"))
	require.NoError(t, store.UpdateStepStatus(ctx, runID, "skipped", "skipped", "ignored-result", ""))
	require.NoError(t, store.CompleteRun(ctx, runID, "failed", "workflow failed"))

	status, err := store.GetRunStatus(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, runID, status.RunID)
	assert.Equal(t, "persisted-status", status.WorkflowName)
	assert.Equal(t, "failed", status.Status)
	assert.Equal(t, 3, status.TotalSteps)
	assert.Equal(t, 3, status.CompletedSteps)
	assert.False(t, status.StartedAt.IsZero())

	stepStatuses := make(map[string]StepStatus, len(status.StepStatuses))
	for _, step := range status.StepStatuses {
		stepStatuses[step.StepID] = step
	}
	require.Len(t, stepStatuses, 3)
	assert.Equal(t, StepStatus{StepID: "done", Agent: "operator", Status: "completed"}, stepStatuses["done"])
	assert.Equal(t, StepStatus{StepID: "failed", Agent: "executor", Status: "failed", Error: "agent exploded"}, stepStatuses["failed"])
	assert.Equal(t, StepStatus{StepID: "skipped", Status: "skipped"}, stepStatuses["skipped"])

	results, err := store.GetStepResults(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"done": "done-result"}, results)

	runUUID := uuid.MustParse(runID)
	run, err := client.WorkflowRun.Get(ctx, runUUID)
	require.NoError(t, err)
	assert.Equal(t, workflowrun.StatusFailed, run.Status)
	assert.Equal(t, "workflow failed", run.ErrorMessage)
	assert.NotNil(t, run.CompletedAt)

	doneStep, err := client.WorkflowStepRun.Query().
		Where(workflowsteprun.RunID(runUUID), workflowsteprun.StepIDEQ("done")).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, "rendered done", doneStep.Prompt)
	assert.Equal(t, workflowsteprun.StatusCompleted, doneStep.Status)
	assert.Equal(t, "done-result", doneStep.Result)
	assert.NotNil(t, doneStep.StartedAt)
	assert.NotNil(t, doneStep.CompletedAt)
}

func TestStateStoreListRunsOrdersByNewestAndHonorsLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, client := newStateStoreCreateRunWithIdPersistsCanonicalRunAndRejectsBadIdStateStore(t)
	base := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	oldID := uuid.New()
	midID := uuid.New()
	newID := uuid.New()

	_, err := client.WorkflowRun.Create().
		SetID(oldID).
		SetWorkflowName("old").
		SetStatus(workflowrun.StatusCompleted).
		SetTotalSteps(1).
		SetCompletedSteps(1).
		SetStartedAt(base).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.WorkflowRun.Create().
		SetID(midID).
		SetWorkflowName("middle").
		SetStatus(workflowrun.StatusRunning).
		SetTotalSteps(2).
		SetCompletedSteps(1).
		SetStartedAt(base.Add(time.Minute)).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.WorkflowRun.Create().
		SetID(newID).
		SetWorkflowName("new").
		SetStatus(workflowrun.StatusPending).
		SetTotalSteps(3).
		SetCompletedSteps(0).
		SetStartedAt(base.Add(2 * time.Minute)).
		Save(ctx)
	require.NoError(t, err)

	runs, err := store.ListRuns(ctx, 2)
	require.NoError(t, err)
	require.Len(t, runs, 2)
	assert.Equal(t, newID.String(), runs[0].RunID)
	assert.Equal(t, "new", runs[0].WorkflowName)
	assert.Equal(t, "pending", runs[0].Status)
	assert.Equal(t, 3, runs[0].TotalSteps)
	assert.Equal(t, 0, runs[0].CompletedSteps)
	assert.Equal(t, base.Add(2*time.Minute), runs[0].StartedAt)
	assert.Equal(t, midID.String(), runs[1].RunID)
	assert.Equal(t, "middle", runs[1].WorkflowName)
	assert.Equal(t, "running", runs[1].Status)
}

func TestStateStoreQueryMethodsRejectMalformedRunIDs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, _ := newStateStoreCreateRunWithIdPersistsCanonicalRunAndRejectsBadIdStateStore(t)

	status, err := store.GetRunStatus(ctx, "bad-run-id")
	require.Error(t, err)
	assert.Nil(t, status)
	assert.ErrorContains(t, err, `parse run ID "bad-run-id"`)

	results, err := store.GetStepResults(ctx, "bad-run-id")
	require.Error(t, err)
	assert.Nil(t, results)
	assert.ErrorContains(t, err, `parse run ID "bad-run-id"`)

	err = store.UpdateRunStatus(ctx, "bad-run-id", "running")
	require.Error(t, err)
	assert.ErrorContains(t, err, `parse run ID "bad-run-id"`)
}
