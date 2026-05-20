package runledger

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ctxkeys"
)

type edgeBranchRunLedgerStore struct {
	*MemoryStore

	appendErr      error
	failAppendCall int
	appendCalls    int

	snapshotErr      error
	failSnapshotCall int
	snapshotCalls    int
	staticSnapshot   *RunSnapshot
}

func (s *edgeBranchRunLedgerStore) AppendJournalEvent(ctx context.Context, event JournalEvent) error {
	s.appendCalls++
	if s.appendErr != nil && s.appendCalls == s.failAppendCall {
		return s.appendErr
	}
	return s.MemoryStore.AppendJournalEvent(ctx, event)
}

func (s *edgeBranchRunLedgerStore) GetRunSnapshot(ctx context.Context, runID string) (*RunSnapshot, error) {
	s.snapshotCalls++
	if s.snapshotErr != nil && s.snapshotCalls == s.failSnapshotCall {
		return nil, s.snapshotErr
	}
	if s.staticSnapshot != nil {
		return s.staticSnapshot.DeepCopy(), nil
	}
	return s.MemoryStore.GetRunSnapshot(ctx, runID)
}

func TestRunCreatePropagatesStoreFailuresBeforeLinking(t *testing.T) {
	testCases := []struct {
		name             string
		store            *edgeBranchRunLedgerStore
		wantErr          string
		wantAppendCalls  int
		wantSnapshotCall int
	}{
		{
			name: "run_created append failure",
			store: &edgeBranchRunLedgerStore{
				MemoryStore:    NewMemoryStore(),
				appendErr:      errors.New("journal unavailable"),
				failAppendCall: 1,
			},
			wantErr:         "append run_created: journal unavailable",
			wantAppendCalls: 1,
		},
		{
			name: "plan_attached append failure",
			store: &edgeBranchRunLedgerStore{
				MemoryStore:    NewMemoryStore(),
				appendErr:      errors.New("tail write refused"),
				failAppendCall: 2,
			},
			wantErr:         "append plan_attached: tail write refused",
			wantAppendCalls: 2,
		},
		{
			name: "snapshot failure",
			store: &edgeBranchRunLedgerStore{
				MemoryStore:      NewMemoryStore(),
				snapshotErr:      errors.New("snapshot projection failed"),
				failSnapshotCall: 1,
			},
			wantErr:          "get snapshot: snapshot projection failed",
			wantAppendCalls:  2,
			wantSnapshotCall: 1,
		},
	}

	planJSON := `{
		"goal": "exercise run create failures",
		"acceptance_criteria": [],
		"steps": [{"id": "s1", "goal": "write tests", "owner_agent": "operator", "validator": {"type": "build_pass"}}]
	}`

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			linker := &stubMissionLinker{}
			tool := buildRunCreate(tc.store, linker)

			result, err := tool.Handler(orchestratorCtx(), map[string]interface{}{
				"plan_json":        planJSON,
				"session_key":      "session-1",
				"original_request": "cover edge branches",
			})

			require.Error(t, err)
			assert.Nil(t, result)
			assert.EqualError(t, err, tc.wantErr)
			assert.Equal(t, tc.wantAppendCalls, tc.store.appendCalls)
			assert.Equal(t, tc.wantSnapshotCall, tc.store.snapshotCalls)
			assert.Zero(t, linker.calls)
		})
	}
}

func TestRunActiveReportsBlockedQueueAndMissingCurrentStep(t *testing.T) {
	ctx := orchestratorCtx()

	t.Run("no executable step", func(t *testing.T) {
		store := &edgeBranchRunLedgerStore{
			MemoryStore: NewMemoryStore(),
			staticSnapshot: &RunSnapshot{
				RunID:  "run-blocked",
				Status: RunStatusRunning,
				Steps: []Step{{
					StepID:    "s1",
					Status:    StepStatusPending,
					DependsOn: []string{"missing"},
				}},
			},
		}

		result, err := buildRunActive(store).Handler(ctx, map[string]interface{}{"run_id": "run-blocked"})

		require.NoError(t, err)
		m := result.(map[string]interface{})
		assert.Equal(t, "no_active_step", m["status"])
		assert.Equal(t, "run-blocked", m["run_id"])
		assert.Contains(t, m["message"], "No step")
	})

	t.Run("current step id not found", func(t *testing.T) {
		store := &edgeBranchRunLedgerStore{
			MemoryStore: NewMemoryStore(),
			staticSnapshot: &RunSnapshot{
				RunID:         "run-missing-current",
				Status:        RunStatusRunning,
				CurrentStepID: "ghost",
				Steps: []Step{{
					StepID: "s1",
					Status: StepStatusPending,
				}},
			},
		}

		result, err := buildRunActive(store).Handler(ctx, map[string]interface{}{"run_id": "run-missing-current"})

		require.NoError(t, err)
		m := result.(map[string]interface{})
		assert.Equal(t, "active", m["status"])
		assert.Nil(t, m["step"])
		assert.Equal(t, RunStatusRunning, m["run_status"])
	})
}

func TestRunApplyPolicyParsesStructuredDecisionAndReportsSnapshotStatus(t *testing.T) {
	ctx := orchestratorCtx()
	store := NewMemoryStore()
	tool := buildRunApplyPolicy(store)

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-policy-structured",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "session-1", Goal: "policy"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-policy-structured",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{
				StepID:     "s1",
				Goal:       "original",
				OwnerAgent: "operator",
				Status:     StepStatusFailed,
				Validator:  ValidatorSpec{Type: ValidatorBuildPass},
				MaxRetries: DefaultMaxRetries,
			}},
		}),
	}))

	result, err := tool.Handler(ctx, map[string]interface{}{
		"run_id":    "run-policy-structured",
		"step_id":   "s1",
		"action":    "decompose",
		"reason":    "split the work",
		"new_agent": "reviewer",
		"new_steps_json": `[{
			"step_id": "s1a",
			"goal": "smaller step",
			"owner_agent": "operator",
			"status": "pending",
			"validator": {"type": "test_pass"},
			"max_retries": 2
		}]`,
		"new_validator_json": `{"type": "command_pass", "target": "go test ./internal/runledger"}`,
	})

	require.NoError(t, err)
	m := result.(map[string]interface{})
	assert.Equal(t, "applied", m["status"])
	assert.Equal(t, "decompose", m["action"])
	assert.Equal(t, RunStatusRunning, m["run_status"])

	events, err := store.GetJournalEvents(ctx, "run-policy-structured")
	require.NoError(t, err)
	require.Len(t, events, 3)
	assert.Equal(t, EventPolicyDecisionApplied, events[2].Type)

	var payload PolicyDecisionAppliedPayload
	require.NoError(t, json.Unmarshal(events[2].Payload, &payload))
	assert.Equal(t, "s1", payload.StepID)
	assert.Equal(t, PolicyDecompose, payload.Decision.Action)
	assert.Equal(t, "split the work", payload.Decision.Reason)
	assert.Equal(t, "reviewer", payload.Decision.NewAgent)
	require.Len(t, payload.Decision.NewSteps, 1)
	assert.Equal(t, "s1a", payload.Decision.NewSteps[0].StepID)
	require.NotNil(t, payload.Decision.NewValidator)
	assert.Equal(t, ValidatorCommandPass, payload.Decision.NewValidator.Type)
}

func TestRunApplyPolicyRejectsInvalidStructuredJSONBeforeJournaling(t *testing.T) {
	testCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name: "invalid new steps",
			params: map[string]interface{}{
				"run_id":         "run-invalid-policy-json",
				"step_id":        "s1",
				"action":         "decompose",
				"reason":         "bad json",
				"new_steps_json": `{"not":"an array"}`,
			},
			wantErr: "parse new_steps_json",
		},
		{
			name: "invalid new validator",
			params: map[string]interface{}{
				"run_id":             "run-invalid-policy-json",
				"step_id":            "s1",
				"action":             "change_validator",
				"reason":             "bad json",
				"new_validator_json": `[`,
			},
			wantErr: "parse new_validator_json",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			store := NewMemoryStore()
			tool := buildRunApplyPolicy(store)

			result, err := tool.Handler(orchestratorCtx(), tc.params)

			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tc.wantErr)
			_, journalErr := store.GetJournalEvents(context.Background(), "run-invalid-policy-json")
			assert.Error(t, journalErr)
		})
	}
}

func TestRunApplyPolicyPropagatesStoreFailures(t *testing.T) {
	testCases := []struct {
		name             string
		store            *edgeBranchRunLedgerStore
		wantErr          string
		wantAppendCalls  int
		wantSnapshotCall int
	}{
		{
			name: "append failure",
			store: &edgeBranchRunLedgerStore{
				MemoryStore:    NewMemoryStore(),
				appendErr:      errors.New("policy journal refused"),
				failAppendCall: 1,
			},
			wantErr:         "append policy_decision_applied: policy journal refused",
			wantAppendCalls: 1,
		},
		{
			name: "snapshot failure",
			store: &edgeBranchRunLedgerStore{
				MemoryStore:      NewMemoryStore(),
				snapshotErr:      errors.New("projection stale"),
				failSnapshotCall: 1,
			},
			wantErr:          "get snapshot: projection stale",
			wantAppendCalls:  1,
			wantSnapshotCall: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := buildRunApplyPolicy(tc.store).Handler(orchestratorCtx(), map[string]interface{}{
				"run_id":  "run-policy-failure",
				"step_id": "s1",
				"action":  "retry",
				"reason":  "retry after failure",
			})

			require.Error(t, err)
			assert.Nil(t, result)
			assert.EqualError(t, err, tc.wantErr)
			assert.Equal(t, tc.wantAppendCalls, tc.store.appendCalls)
			assert.Equal(t, tc.wantSnapshotCall, tc.store.snapshotCalls)
		})
	}
}

func TestRunResumeRejectsNonPausedRunWithoutAppending(t *testing.T) {
	ctx := orchestratorCtx()
	store := NewMemoryStore()
	tool := buildRunResume(store)

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-active",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "session-1", Goal: "active"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-active",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{
				StepID:     "s1",
				Goal:       "work",
				OwnerAgent: "operator",
				Status:     StepStatusPending,
				Validator:  ValidatorSpec{Type: ValidatorBuildPass},
				MaxRetries: DefaultMaxRetries,
			}},
		}),
	}))
	before, err := store.GetJournalEvents(ctx, "run-active")
	require.NoError(t, err)

	result, err := tool.Handler(ctx, map[string]interface{}{"run_id": "run-active"})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrRunNotPaused)
	after, err := store.GetJournalEvents(ctx, "run-active")
	require.NoError(t, err)
	assert.Len(t, after, len(before))
}

func TestRunResumeAppendsResumedByAndReturnsSummary(t *testing.T) {
	ctx := ctxkeys.WithAgentName(context.Background(), "lango-orchestrator")
	store := NewMemoryStore()
	tool := buildRunResume(store)

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-paused",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "session-1", Goal: "paused"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: "run-paused",
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{
				StepID:     "s1",
				Goal:       "work",
				OwnerAgent: "operator",
				Status:     StepStatusPending,
				Validator:  ValidatorSpec{Type: ValidatorBuildPass},
				MaxRetries: DefaultMaxRetries,
			}},
		}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-paused",
		Type:    EventRunPaused,
		Payload: marshalPayload(RunPausedPayload{Reason: "turn limit"}),
	}))

	result, err := tool.Handler(ctx, map[string]interface{}{"run_id": "run-paused"})

	require.NoError(t, err)
	m := result.(map[string]interface{})
	assert.Equal(t, "resumed", m["status"])
	assert.Equal(t, "run-paused", m["run_id"])
	require.NotNil(t, m["summary"])
	summary := m["summary"].(RunSummary)
	assert.Equal(t, RunStatusRunning, summary.Status)

	events, err := store.GetJournalEvents(ctx, "run-paused")
	require.NoError(t, err)
	require.Len(t, events, 4)
	assert.Equal(t, EventRunResumed, events[3].Type)
	var payload RunResumedPayload
	require.NoError(t, json.Unmarshal(events[3].Payload, &payload))
	assert.Equal(t, "lango-orchestrator", payload.ResumedBy)
}
