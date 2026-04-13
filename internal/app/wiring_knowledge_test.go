package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/runledger"
)

func TestRunSummaryProviderAdapter_FiltersTerminalRuns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := runledger.NewMemoryStore()

	seed := func(runID string, status runledger.RunStatus) {
		require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
			RunID:   runID,
			Type:    runledger.EventRunCreated,
			Payload: mustJSON(t, runledger.RunCreatedPayload{SessionKey: "sess-1", Goal: runID}),
		}))
		require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
			RunID: runID,
			Type:  runledger.EventPlanAttached,
			Payload: mustJSON(t, runledger.PlanAttachedPayload{
				Steps: []runledger.Step{{
					StepID:     "s1",
					Goal:       "work",
					OwnerAgent: "operator",
					Status:     runledger.StepStatusPending,
					Validator:  runledger.ValidatorSpec{Type: runledger.ValidatorBuildPass},
					MaxRetries: runledger.DefaultMaxRetries,
				}},
			}),
		}))
		switch status {
		case runledger.RunStatusPaused:
			require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
				RunID:   runID,
				Type:    runledger.EventRunPaused,
				Payload: mustJSON(t, runledger.RunPausedPayload{Reason: "paused"}),
			}))
		case runledger.RunStatusCompleted:
			require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
				RunID:   runID,
				Type:    runledger.EventRunCompleted,
				Payload: mustJSON(t, runledger.RunCompletedPayload{Summary: "done"}),
			}))
		case runledger.RunStatusFailed:
			require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
				RunID:   runID,
				Type:    runledger.EventRunFailed,
				Payload: mustJSON(t, runledger.RunFailedPayload{Reason: "failed"}),
			}))
		}
	}

	seed("run-running", runledger.RunStatusRunning)
	seed("run-paused", runledger.RunStatusPaused)
	seed("run-completed", runledger.RunStatusCompleted)
	seed("run-failed", runledger.RunStatusFailed)

	adapter := &runSummaryProviderAdapter{store: store}
	summaries, err := adapter.ListRunSummaries(ctx, "sess-1", 10)
	require.NoError(t, err)
	require.Len(t, summaries, 2)

	assert.ElementsMatch(t, []string{"run-running", "run-paused"}, []string{
		summaries[0].RunID,
		summaries[1].RunID,
	})
}

func mustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}
