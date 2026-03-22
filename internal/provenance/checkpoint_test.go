package provenance

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/runledger"
)

// testPayload marshals v to json.RawMessage for test journal events.
func testPayload(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func TestCheckpointService_CreateManual(t *testing.T) {
	store := NewMemoryStore()
	ledger := runledger.NewMemoryStore()
	cfg := config.CheckpointConfig{
		AutoOnStepComplete: true,
		AutoOnPolicy:       true,
		MaxPerSession:      100,
	}

	ctx := context.Background()
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-1",
		Type:    runledger.EventRunCreated,
		Payload: testPayload(runledger.RunCreatedPayload{SessionKey: "sess-1", Goal: "test"}),
	}))

	svc := NewCheckpointService(store, ledger, cfg)

	cp, err := svc.CreateManual(ctx, "sess-1", "run-1", "my checkpoint")
	require.NoError(t, err)
	assert.NotEmpty(t, cp.ID)
	assert.Equal(t, "sess-1", cp.SessionKey)
	assert.Equal(t, "run-1", cp.RunID)
	assert.Equal(t, "my checkpoint", cp.Label)
	assert.Equal(t, TriggerManual, cp.Trigger)
	assert.Equal(t, int64(1), cp.JournalSeq)
}

func TestCheckpointService_CreateManual_EmptyLabel(t *testing.T) {
	store := NewMemoryStore()
	cfg := config.CheckpointConfig{}
	svc := NewCheckpointService(store, nil, cfg)

	_, err := svc.CreateManual(context.Background(), "sess-1", "run-1", "")
	assert.ErrorIs(t, err, ErrInvalidLabel)
}

func TestCheckpointService_CreateManual_EmptyRunID(t *testing.T) {
	store := NewMemoryStore()
	cfg := config.CheckpointConfig{}
	svc := NewCheckpointService(store, nil, cfg)

	_, err := svc.CreateManual(context.Background(), "sess-1", "", "label")
	assert.ErrorIs(t, err, ErrInvalidRunID)
}

func TestCheckpointService_MaxCheckpoints(t *testing.T) {
	store := NewMemoryStore()
	ledger := runledger.NewMemoryStore()
	cfg := config.CheckpointConfig{MaxPerSession: 2}

	ctx := context.Background()
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-1",
		Type:    runledger.EventRunCreated,
		Payload: testPayload(runledger.RunCreatedPayload{SessionKey: "sess-1", Goal: "test"}),
	}))

	svc := NewCheckpointService(store, ledger, cfg)

	_, err := svc.CreateManual(ctx, "sess-1", "run-1", "cp-1")
	require.NoError(t, err)

	_, err = svc.CreateManual(ctx, "sess-1", "run-1", "cp-2")
	require.NoError(t, err)

	_, err = svc.CreateManual(ctx, "sess-1", "run-1", "cp-3")
	assert.ErrorIs(t, err, ErrMaxCheckpoints)
}

func TestCheckpointService_OnJournalEvent_StepComplete(t *testing.T) {
	store := NewMemoryStore()
	ledger := runledger.NewMemoryStore()
	cfg := config.CheckpointConfig{
		AutoOnStepComplete: true,
		MaxPerSession:      100,
	}

	ctx := context.Background()
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-1",
		Type:    runledger.EventRunCreated,
		Payload: testPayload(runledger.RunCreatedPayload{SessionKey: "sess-1", Goal: "test"}),
	}))

	svc := NewCheckpointService(store, ledger, cfg)

	svc.OnJournalEvent(runledger.JournalEvent{
		RunID: "run-1",
		Seq:   2,
		Type:  runledger.EventStepValidationPassed,
	})

	list, err := store.ListByRun(ctx, "run-1")
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, TriggerStepComplete, list[0].Trigger)
}

func TestCheckpointService_OnJournalEvent_Disabled(t *testing.T) {
	store := NewMemoryStore()
	cfg := config.CheckpointConfig{
		AutoOnStepComplete: false,
		AutoOnPolicy:       false,
	}

	svc := NewCheckpointService(store, nil, cfg)

	svc.OnJournalEvent(runledger.JournalEvent{
		RunID: "run-1",
		Seq:   1,
		Type:  runledger.EventStepValidationPassed,
	})

	list, err := store.ListByRun(context.Background(), "run-1")
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestAppendHook_Integration(t *testing.T) {
	cpStore := NewMemoryStore()
	cfg := config.CheckpointConfig{
		AutoOnStepComplete: true,
		MaxPerSession:      100,
	}

	var svc *CheckpointService
	ledger := runledger.NewMemoryStore(runledger.WithAppendHook(func(event runledger.JournalEvent) {
		if svc != nil {
			svc.OnJournalEvent(event)
		}
	}))

	svc = NewCheckpointService(cpStore, ledger, cfg)

	ctx := context.Background()
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-1",
		Type:    runledger.EventRunCreated,
		Payload: testPayload(runledger.RunCreatedPayload{SessionKey: "sess-1", Goal: "test"}),
	}))

	// This append should trigger a checkpoint via the hook.
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID: "run-1",
		Type:  runledger.EventStepValidationPassed,
		Payload: testPayload(runledger.StepValidationPassedPayload{
			StepID: "step-1",
			Result: runledger.ValidationResult{Passed: true},
		}),
	}))

	list, err := cpStore.ListByRun(ctx, "run-1")
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, TriggerStepComplete, list[0].Trigger)
}

func TestSetAppendHook_Integration(t *testing.T) {
	// Simulates the real app wiring: ledger store is created first,
	// then provenance module registers a hook via SetAppendHook.
	cpStore := NewMemoryStore()
	cfg := config.CheckpointConfig{
		AutoOnStepComplete: true,
		MaxPerSession:      100,
	}

	ledger := runledger.NewMemoryStore()
	svc := NewCheckpointService(cpStore, ledger, cfg)

	// Post-construction hook registration — mirrors modules_provenance.go wiring.
	ledger.SetAppendHook(svc.OnJournalEvent)

	ctx := context.Background()
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-1",
		Type:    runledger.EventRunCreated,
		Payload: testPayload(runledger.RunCreatedPayload{SessionKey: "sess-1", Goal: "test"}),
	}))

	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID: "run-1",
		Type:  runledger.EventStepValidationPassed,
		Payload: testPayload(runledger.StepValidationPassedPayload{
			StepID: "step-1",
			Result: runledger.ValidationResult{Passed: true},
		}),
	}))

	list, err := cpStore.ListByRun(ctx, "run-1")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, TriggerStepComplete, list[0].Trigger)
	assert.Contains(t, list[0].Label, "step_validated_2")
}
