package runledger

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/enttest"
)

func TestEntStore_JournalAndSnapshotRoundTrip(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	store := NewEntStore(client)
	ctx := context.Background()

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "s1", Goal: "goal"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-1",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{
				StepID:     "step-1",
				Goal:       "do work",
				OwnerAgent: "operator",
				Status:     StepStatusPending,
				Validator:  ValidatorSpec{Type: ValidatorBuildPass},
				MaxRetries: DefaultMaxRetries,
			}},
		}),
	}))

	snap, err := store.GetRunSnapshot(ctx, "run-1")
	require.NoError(t, err)
	require.NoError(t, store.UpdateCachedSnapshot(ctx, snap))

	reloaded, err := store.GetRunSnapshot(ctx, "run-1")
	require.NoError(t, err)
	assert.Equal(t, "run-1", reloaded.RunID)
	assert.Equal(t, "goal", reloaded.Goal)
	require.Len(t, reloaded.Steps, 1)
	assert.Equal(t, "step-1", reloaded.Steps[0].StepID)
}

func TestEntStore_ListRunsUsesPersistentSnapshots(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	store := NewEntStore(client)
	ctx := context.Background()

	for _, runID := range []string{"run-a", "run-b"} {
		require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
			RunID:   runID,
			Type:    EventRunCreated,
			Payload: marshalPayload(RunCreatedPayload{SessionKey: "s", Goal: runID}),
		}))
		require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
			RunID: runID,
			Type:  EventPlanAttached,
			Payload: marshalPayload(PlanAttachedPayload{
				Steps: []Step{{
					StepID:     "s1",
					Goal:       "g",
					OwnerAgent: "operator",
					Status:     StepStatusPending,
					Validator:  ValidatorSpec{Type: ValidatorBuildPass},
					MaxRetries: DefaultMaxRetries,
				}},
			}),
		}))
		snap, err := store.GetRunSnapshot(ctx, runID)
		require.NoError(t, err)
		require.NoError(t, store.UpdateCachedSnapshot(ctx, snap))
	}

	runs, err := store.ListRuns(ctx, 10)
	require.NoError(t, err)
	require.Len(t, runs, 2)
	assert.ElementsMatch(t, []string{"run-a", "run-b"}, []string{runs[0].RunID, runs[1].RunID})
}
