package workflow

import (
	"context"
	"encoding/json"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/runledger"
)

func TestMaybeListRunsFromLedger(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	store := runledger.NewEntStore(client)
	ctx := context.Background()
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-1",
		Type:    runledger.EventRunCreated,
		Payload: runledgerMarshal(runledger.RunCreatedPayload{SessionKey: "s1", Goal: "wf-a"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID: "run-1",
		Type:  runledger.EventPlanAttached,
		Payload: runledgerMarshal(runledger.PlanAttachedPayload{
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
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.AuthoritativeRead = true
	boot := &bootstrap.Result{Config: cfg, DBClient: client}

	runs, handled, err := maybeListRunsFromLedger(boot, 10)
	require.NoError(t, err)
	assert.True(t, handled)
	require.Len(t, runs, 1)
	assert.Equal(t, "wf-a", runs[0].WorkflowName)
}

func TestMaybeStatusFromLedger(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	store := runledger.NewEntStore(client)
	ctx := context.Background()
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-2",
		Type:    runledger.EventRunCreated,
		Payload: runledgerMarshal(runledger.RunCreatedPayload{SessionKey: "s1", Goal: "wf-b"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID: "run-2",
		Type:  runledger.EventPlanAttached,
		Payload: runledgerMarshal(runledger.PlanAttachedPayload{
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

	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.AuthoritativeRead = true
	boot := &bootstrap.Result{Config: cfg, DBClient: client}

	status, handled, err := maybeStatusFromLedger(boot, "run-2")
	require.NoError(t, err)
	assert.True(t, handled)
	require.NotNil(t, status)
	assert.Equal(t, "wf-b", status.WorkflowName)
	require.Len(t, status.StepStatuses, 1)
	assert.Equal(t, "step-1", status.StepStatuses[0].StepID)
}

func runledgerMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
