package runledger

import (
	"context"
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
}
