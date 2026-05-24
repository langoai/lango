package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/storage"
)

func executeWorkflowCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func workflowLedgerBootLoader(t *testing.T) func() (*bootstrap.Result, error) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:workflow-output?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	store := runledger.NewEntStore(client)
	ctx := context.Background()
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-1",
		Type:    runledger.EventRunCreated,
		Payload: workflowRunLedgerMarshal(runledger.RunCreatedPayload{SessionKey: "s1", Goal: "wf-a"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID: "run-1",
		Type:  runledger.EventPlanAttached,
		Payload: workflowRunLedgerMarshal(runledger.PlanAttachedPayload{
			Steps: []runledger.Step{{
				StepID:     "step-1",
				Goal:       "work",
				OwnerAgent: "operator",
				Status:     runledger.StepStatusCompleted,
				Validator:  runledger.ValidatorSpec{Type: runledger.ValidatorBuildPass},
				MaxRetries: runledger.DefaultMaxRetries,
			}},
		}),
	}))
	snap, err := store.GetRunSnapshot(ctx, "run-1")
	require.NoError(t, err)
	require.NoError(t, store.UpdateCachedSnapshot(ctx, snap))

	cfg := config.DefaultConfig()
	cfg.Workflow.Enabled = true
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.AuthoritativeRead = true
	facade := storage.NewFacade(nil, nil, storage.WithEntClient(client))

	return func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:  cfg,
			Storage: facade,
		}, nil
	}
}

func workflowRunLedgerMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

func TestWorkflowList_WritesTableToCommandWriter(t *testing.T) {
	cmd := newWorkflowListCmd(workflowLedgerBootLoader(t))
	out, err := executeWorkflowCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "wf-a")
}

func TestWorkflowStatus_WritesDetailsToCommandWriter(t *testing.T) {
	cmd := newWorkflowStatusCmd(workflowLedgerBootLoader(t))
	out, err := executeWorkflowCmd(t, cmd, "run-1")
	require.NoError(t, err)
	assert.Contains(t, out, "Run ID:    run-1")
	assert.Contains(t, out, "Workflow:  wf-a")
	assert.Contains(t, out, "step-1")
}

func TestWorkflowHistory_WritesTableToCommandWriter(t *testing.T) {
	cmd := newWorkflowHistoryCmd(workflowLedgerBootLoader(t))
	out, err := executeWorkflowCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "wf-a")
}

func TestWorkflowCancel_WritesSuccessToCommandWriter(t *testing.T) {
	original := cancelWorkflowRun
	cancelWorkflowRun = func(_ func() (*bootstrap.Result, error), runID string) (string, error) {
		return "Workflow run " + runID + " cancelled.", nil
	}
	t.Cleanup(func() { cancelWorkflowRun = original })

	cmd := newWorkflowCancelCmd(func() (*bootstrap.Result, error) { return nil, assert.AnError })
	out, err := executeWorkflowCmd(t, cmd, "run-1")
	require.NoError(t, err)
	assert.Contains(t, out, "Workflow run run-1 cancelled.")
}
