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

// orchestratorCtx returns a context that identifies the caller as the orchestrator.
func orchestratorCtx() context.Context {
	return context.Background() // empty agent name = orchestrator
}

// executionCtx returns a context that identifies the caller as an execution agent.
func executionCtx() context.Context {
	return ctxkeys.WithAgentName(context.Background(), "operator")
}

func TestBuildTools_Count(t *testing.T) {
	store := NewMemoryStore()
	pev := NewPEVEngine(store, DefaultValidators())
	tools := BuildTools(store, pev)
	assert.Len(t, tools, 8)

	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.Name] = true
	}
	assert.True(t, names["run_create"])
	assert.True(t, names["run_read"])
	assert.True(t, names["run_active"])
	assert.True(t, names["run_note"])
	assert.True(t, names["run_propose_step_result"])
	assert.True(t, names["run_apply_policy"])
	assert.True(t, names["run_approve_step"])
	assert.True(t, names["run_resume"])
}

// toolMap builds a map of tool name -> handler for convenient test calls.
func toolMap(store *MemoryStore, pev *PEVEngine) map[string]*runCreateHelper {
	tools := BuildTools(store, pev)
	m := make(map[string]*runCreateHelper, len(tools))
	for _, tool := range tools {
		m[tool.Name] = &runCreateHelper{tool.Handler}
	}
	return m
}

func TestRunCreate_EndToEnd(t *testing.T) {
	ctx := orchestratorCtx()
	execCtx := executionCtx()
	store := NewMemoryStore()

	// Use a mock validator that always passes for PEV auto-verification.
	mockValidators := map[ValidatorType]Validator{
		ValidatorBuildPass:            &mockValidator{result: &ValidationResult{Passed: true, Reason: "build ok"}},
		ValidatorTestPass:             &mockValidator{result: &ValidationResult{Passed: true, Reason: "tests ok"}},
		ValidatorOrchestratorApproval: &OrchestratorApprovalValidator{},
	}
	pev := NewPEVEngine(store, mockValidators)
	tm := toolMap(store, pev)

	// Create a run with a valid plan.
	planJSON := `{
		"goal": "implement Task OS",
		"acceptance_criteria": [
			{"description": "build passes", "validator": {"type": "build_pass", "target": "./..."}}
		],
		"steps": [
			{"id": "s1", "goal": "write code", "owner_agent": "operator", "validator": {"type": "build_pass"}},
			{"id": "s2", "goal": "test code", "owner_agent": "operator", "validator": {"type": "test_pass"}, "depends_on": ["s1"]}
		]
	}`

	result, err := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json":        planJSON,
		"session_key":      "session-1",
		"original_request": "Build the Task OS",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "created", m["status"])
	runID := m["run_id"].(string)
	assert.NotEmpty(t, runID)
	assert.Equal(t, 2, m["step_count"])

	// Read the run.
	readResult, err := tm["run_read"].call(ctx, map[string]interface{}{"run_id": runID})
	require.NoError(t, err)
	snap := readResult.(*RunSnapshot)
	assert.Equal(t, RunStatusRunning, snap.Status)
	assert.Len(t, snap.Steps, 2)

	// Get active step.
	activeResult, err := tm["run_active"].call(ctx, map[string]interface{}{"run_id": runID})
	require.NoError(t, err)
	am := activeResult.(map[string]interface{})
	assert.Equal(t, "next_available", am["status"])

	// Write a note.
	_, err = tm["run_note"].call(ctx, map[string]interface{}{
		"run_id": runID,
		"key":    "scratch",
		"value":  "testing in progress",
	})
	require.NoError(t, err)

	// Read note back.
	noteResult, err := tm["run_note"].call(ctx, map[string]interface{}{
		"run_id": runID,
		"key":    "scratch",
	})
	require.NoError(t, err)
	nm := noteResult.(map[string]interface{})
	assert.Equal(t, "testing in progress", nm["value"])

	// Start step s1, then propose result (as execution agent).
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})

	proposeResult, err := tm["run_propose_step_result"].call(execCtx, map[string]interface{}{
		"run_id":  runID,
		"step_id": "s1",
		"result":  "code written successfully",
	})
	require.NoError(t, err)

	// With mock pass validator, PEV auto-verifies: step completed, run still in progress.
	pm := proposeResult.(map[string]interface{})
	assert.Equal(t, "verified", pm["status"])
	assert.Equal(t, "running", pm["run_status"])

	snap2, err := store.GetRunSnapshot(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, StepStatusCompleted, snap2.Steps[0].Status)
}

func TestRunCreate_InvalidPlan(t *testing.T) {
	ctx := orchestratorCtx()
	store := NewMemoryStore()
	pev := NewPEVEngine(store, DefaultValidators())
	tm := toolMap(store, pev)

	// Malformed JSON.
	result, err := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json":        "not json",
		"session_key":      "s1",
		"original_request": "test",
	})
	require.NoError(t, err)
	m := result.(map[string]interface{})
	assert.Equal(t, "parse_failed", m["error"])

	// Valid JSON but fails validation.
	result2, err := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json":        `{"goal": "", "steps": []}`,
		"session_key":      "s1",
		"original_request": "test",
	})
	require.NoError(t, err)
	m2 := result2.(map[string]interface{})
	assert.Equal(t, "validation_failed", m2["error"])
}

func TestRunApplyPolicy_Retry(t *testing.T) {
	ctx := orchestratorCtx()
	store := NewMemoryStore()
	pev := NewPEVEngine(store, DefaultValidators())
	tm := toolMap(store, pev)

	planJSON := `{
		"goal": "test retry",
		"acceptance_criteria": [],
		"steps": [{"id": "s1", "goal": "do", "owner_agent": "op", "validator": {"type": "build_pass"}}]
	}`
	res, _ := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json": planJSON, "session_key": "s1", "original_request": "test",
	})
	runID := res.(map[string]interface{})["run_id"].(string)

	// Mark step as started then failed via journal.
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})
	_ = store.RecordValidationResult(ctx, runID, "s1", ValidationResult{Passed: false, Reason: "build failed"})

	// Apply retry policy.
	res2, err := tm["run_apply_policy"].call(ctx, map[string]interface{}{
		"run_id":  runID,
		"step_id": "s1",
		"action":  "retry",
		"reason":  "transient failure",
	})
	require.NoError(t, err)
	m := res2.(map[string]interface{})
	assert.Equal(t, "applied", m["status"])

	// Verify step is back to pending with incremented retry count.
	snap, _ := store.GetRunSnapshot(ctx, runID)
	assert.Equal(t, StepStatusPending, snap.Steps[0].Status)
	assert.Equal(t, 1, snap.Steps[0].RetryCount)
}

func TestResumeIntentDetection(t *testing.T) {
	tests := []struct {
		give string
		want bool
	}{
		{"계속해줘", true},
		{"이어서 작업해줘", true},
		{"please resume the task", true},
		{"continue where we left off", true},
		{"다시 시작해줘", true},
		{"마저 해줘", true},
		{"hello there", false},
		{"build a new feature", false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, DetectResumeIntent(tt.give), "input: %q", tt.give)
	}
}

// runCreateHelper wraps a tool handler for concise test calls.
type runCreateHelper struct {
	handler func(context.Context, map[string]interface{}) (interface{}, error)
}

func (h *runCreateHelper) call(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return h.handler(ctx, params)
}

func TestRunApproveStep(t *testing.T) {
	ctx := orchestratorCtx()
	execCtx := executionCtx()
	store := NewMemoryStore()

	// orchestrator_approval never auto-passes, so PEV returns verification_failed.
	pev := NewPEVEngine(store, DefaultValidators())
	tm := toolMap(store, pev)

	planJSON := `{
		"goal": "test approval",
		"acceptance_criteria": [],
		"steps": [{"id": "s1", "goal": "review", "owner_agent": "op", "validator": {"type": "orchestrator_approval"}}]
	}`
	res, _ := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json": planJSON, "session_key": "s1", "original_request": "test",
	})
	runID := res.(map[string]interface{})["run_id"].(string)

	// Start step and propose result (as execution agent).
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})

	// Propose result — PEV auto-runs orchestrator_approval validator which always fails.
	proposeResult, err := tm["run_propose_step_result"].call(execCtx, map[string]interface{}{
		"run_id":  runID,
		"step_id": "s1",
		"result":  "work done",
	})
	require.NoError(t, err)
	pm := proposeResult.(map[string]interface{})
	assert.Equal(t, "verification_failed", pm["status"])
	assert.Contains(t, pm["failure_reason"], "awaiting orchestrator approval")

	// Step should now be in verify_pending (propose event) then failed (validation failed).
	// Actually after PEV records failed validation, step is "failed".
	// But for run_approve_step we need verify_pending.
	// The orchestrator_approval flow: propose -> verify_pending -> validation_failed -> failed.
	// The approve tool should work on orchestrator_approval steps that are in failed state after PEV.
	// Wait — re-reading the plan: approve should only work on verify_pending steps.
	// But PEV already transitions to failed... Let me check the actual state.
	snap, _ := store.GetRunSnapshot(ctx, runID)
	// After PEV validation_failed, step is "failed".
	assert.Equal(t, StepStatusFailed, snap.Steps[0].Status)

	// For orchestrator_approval, the flow is:
	// propose -> verify_pending -> PEV auto-verify -> validation_failed -> step failed
	// -> orchestrator applies policy (retry or approve)
	// The approve tool expects verify_pending, but PEV already moved to failed.
	// We need to retry first to get back to pending, then re-propose.
	// OR: We use run_apply_policy with action=retry, then start again.
	// Actually the design says: for orchestrator_approval, the correct flow is:
	// 1. propose -> PEV runs OrchestratorApprovalValidator -> always fails -> verification_failed returned
	// 2. Orchestrator sees "awaiting orchestrator approval" and calls run_approve_step
	// But run_approve_step now checks verify_pending status...
	// The issue: PEV records validation_failed which moves step to "failed".
	// We need to change: for orchestrator_approval, the approve path should allow "failed" status too,
	// OR PEV should not record validation_failed for orchestrator_approval.
	// Actually looking at the plan more carefully:
	// "orchestrator_approval 특수 처리: PEV가 OrchestratorApprovalValidator를 실행하면 항상 failed → verification_failed"
	// The step goes to failed, then orchestrator uses run_approve_step.
	// So run_approve_step should accept failed status for orchestrator_approval steps.
	// Let me adjust the test to match the actual flow.
}

func TestRunApproveStep_OrchestratorApproval(t *testing.T) {
	ctx := orchestratorCtx()
	execCtx := executionCtx()
	store := NewMemoryStore()
	pev := NewPEVEngine(store, DefaultValidators())
	tm := toolMap(store, pev)

	planJSON := `{
		"goal": "test approval",
		"acceptance_criteria": [],
		"steps": [{"id": "s1", "goal": "review", "owner_agent": "op", "validator": {"type": "orchestrator_approval"}}]
	}`
	res, _ := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json": planJSON, "session_key": "s1", "original_request": "test",
	})
	runID := res.(map[string]interface{})["run_id"].(string)

	// Start step.
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})

	// Propose result -> PEV auto-runs -> orchestrator_approval always fails.
	proposeResult, err := tm["run_propose_step_result"].call(execCtx, map[string]interface{}{
		"run_id": runID, "step_id": "s1", "result": "work done",
	})
	require.NoError(t, err)
	pm := proposeResult.(map[string]interface{})
	assert.Equal(t, "verification_failed", pm["status"])

	// Step is now "failed" after PEV. Orchestrator approves it directly.
	approveRes, err := tm["run_approve_step"].call(ctx, map[string]interface{}{
		"run_id": runID, "step_id": "s1", "reason": "looks good",
	})
	require.NoError(t, err)
	am := approveRes.(map[string]interface{})
	assert.Equal(t, "approved", am["status"])

	snap, _ := store.GetRunSnapshot(ctx, runID)
	assert.Equal(t, StepStatusCompleted, snap.Steps[0].Status)
}

func TestPlannerOutputJSON_Roundtrip(t *testing.T) {
	plan := PlannerOutput{
		Goal: "build feature",
		Steps: []StepInput{
			{
				ID:         "s1",
				Goal:       "write code",
				OwnerAgent: "operator",
				Validator:  ValidatorSpec{Type: ValidatorBuildPass, Target: "./..."},
			},
		},
		AcceptanceCriteria: []AcceptanceCriteriaInput{
			{Description: "tests pass", Validator: ValidatorSpec{Type: ValidatorTestPass}},
		},
	}

	data, err := json.Marshal(plan)
	require.NoError(t, err)

	parsed, err := ParsePlannerOutput(string(data))
	require.NoError(t, err)
	assert.Equal(t, plan.Goal, parsed.Goal)
	assert.Len(t, parsed.Steps, 1)
}

// --- Fix 1 tests: Access control ---

func TestCheckRole_OrchestratorBlockedFromExecutionTools(t *testing.T) {
	ctx := orchestratorCtx()
	err := checkRole(ctx, roleExecution)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrAccessDenied))
}

func TestCheckRole_ExecutionAgentBlockedFromOrchestratorTools(t *testing.T) {
	ctx := executionCtx()
	err := checkRole(ctx, roleOrchestrator)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrAccessDenied))
}

func TestCheckRole_ExecutionAgentAllowedForExecutionTools(t *testing.T) {
	ctx := executionCtx()
	err := checkRole(ctx, roleExecution)
	require.NoError(t, err)
}

func TestProposeStepResult_OrchestratorBlocked(t *testing.T) {
	ctx := orchestratorCtx()
	store := NewMemoryStore()
	pev := NewPEVEngine(store, DefaultValidators())
	tm := toolMap(store, pev)

	_, err := tm["run_propose_step_result"].call(ctx, map[string]interface{}{
		"run_id": "any", "step_id": "any", "result": "test",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrAccessDenied))
}

// --- Fix 2 tests: run_approve_step validator type validation ---

func TestApproveStep_RejectNonOrchestratorApprovalType(t *testing.T) {
	ctx := orchestratorCtx()
	execCtx := executionCtx()
	store := NewMemoryStore()
	mockValidators := map[ValidatorType]Validator{
		ValidatorBuildPass: &mockValidator{result: &ValidationResult{Passed: false, Reason: "build failed"}},
	}
	pev := NewPEVEngine(store, mockValidators)
	tm := toolMap(store, pev)

	planJSON := `{
		"goal": "test",
		"acceptance_criteria": [],
		"steps": [{"id": "s1", "goal": "build", "owner_agent": "op", "validator": {"type": "build_pass"}}]
	}`
	res, _ := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json": planJSON, "session_key": "s1", "original_request": "test",
	})
	runID := res.(map[string]interface{})["run_id"].(string)

	// Start and propose (gets to verify_pending via journal, then PEV fails it).
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID, Type: EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})
	_, _ = tm["run_propose_step_result"].call(execCtx, map[string]interface{}{
		"run_id": runID, "step_id": "s1", "result": "done",
	})

	// Try to approve a build_pass step -> should fail.
	_, err := tm["run_approve_step"].call(ctx, map[string]interface{}{
		"run_id": runID, "step_id": "s1",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrAccessDenied))
	assert.Contains(t, err.Error(), "orchestrator_approval")
}

func TestApproveStep_RejectWrongStatus(t *testing.T) {
	ctx := orchestratorCtx()
	store := NewMemoryStore()
	pev := NewPEVEngine(store, DefaultValidators())
	tm := toolMap(store, pev)

	planJSON := `{
		"goal": "test",
		"acceptance_criteria": [],
		"steps": [{"id": "s1", "goal": "review", "owner_agent": "op", "validator": {"type": "orchestrator_approval"}}]
	}`
	res, _ := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json": planJSON, "session_key": "s1", "original_request": "test",
	})
	runID := res.(map[string]interface{})["run_id"].(string)

	// Step is "pending" — not verify_pending or failed.
	_, err := tm["run_approve_step"].call(ctx, map[string]interface{}{
		"run_id": runID, "step_id": "s1",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected verify_pending")
}

// --- Fix 4 tests: PEV auto-verification + run completion ---

func TestProposeResult_AutoVerify_Pass(t *testing.T) {
	ctx := orchestratorCtx()
	execCtx := executionCtx()
	store := NewMemoryStore()
	mockValidators := map[ValidatorType]Validator{
		ValidatorBuildPass: &mockValidator{result: &ValidationResult{Passed: true, Reason: "ok"}},
	}
	pev := NewPEVEngine(store, mockValidators)
	tm := toolMap(store, pev)

	planJSON := `{
		"goal": "test verify",
		"acceptance_criteria": [],
		"steps": [
			{"id": "s1", "goal": "build", "owner_agent": "op", "validator": {"type": "build_pass"}},
			{"id": "s2", "goal": "build more", "owner_agent": "op", "validator": {"type": "build_pass"}, "depends_on": ["s1"]}
		]
	}`
	res, _ := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json": planJSON, "session_key": "s1", "original_request": "test",
	})
	runID := res.(map[string]interface{})["run_id"].(string)

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID, Type: EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})

	result, err := tm["run_propose_step_result"].call(execCtx, map[string]interface{}{
		"run_id": runID, "step_id": "s1", "result": "done",
	})
	require.NoError(t, err)
	m := result.(map[string]interface{})
	assert.Equal(t, "verified", m["status"])
	assert.Equal(t, "running", m["run_status"]) // s2 still pending

	snap, _ := store.GetRunSnapshot(ctx, runID)
	assert.Equal(t, StepStatusCompleted, snap.Steps[0].Status)
}

func TestProposeResult_AutoVerify_RunCompletion(t *testing.T) {
	ctx := orchestratorCtx()
	execCtx := executionCtx()
	store := NewMemoryStore()
	mockValidators := map[ValidatorType]Validator{
		ValidatorBuildPass: &mockValidator{result: &ValidationResult{Passed: true, Reason: "ok"}},
	}
	pev := NewPEVEngine(store, mockValidators)
	tm := toolMap(store, pev)

	planJSON := `{
		"goal": "single step",
		"acceptance_criteria": [
			{"description": "build ok", "validator": {"type": "build_pass"}}
		],
		"steps": [{"id": "s1", "goal": "build", "owner_agent": "op", "validator": {"type": "build_pass"}}]
	}`
	res, _ := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json": planJSON, "session_key": "s1", "original_request": "test",
	})
	runID := res.(map[string]interface{})["run_id"].(string)

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID, Type: EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})

	result, err := tm["run_propose_step_result"].call(execCtx, map[string]interface{}{
		"run_id": runID, "step_id": "s1", "result": "done",
	})
	require.NoError(t, err)
	m := result.(map[string]interface{})
	assert.Equal(t, "verified", m["status"])
	assert.Equal(t, "completed", m["run_status"])

	snap, _ := store.GetRunSnapshot(ctx, runID)
	assert.Equal(t, RunStatusCompleted, snap.Status)
}

func TestProposeResult_AutoVerify_CriteriaUnmet(t *testing.T) {
	ctx := orchestratorCtx()
	execCtx := executionCtx()
	store := NewMemoryStore()

	// Step validator passes, but acceptance criteria validator fails.
	mockValidators := map[ValidatorType]Validator{
		ValidatorBuildPass: &mockValidator{result: &ValidationResult{Passed: true, Reason: "ok"}},
		ValidatorTestPass:  &mockValidator{result: &ValidationResult{Passed: false, Reason: "tests fail"}},
	}
	pev := NewPEVEngine(store, mockValidators)
	tm := toolMap(store, pev)

	planJSON := `{
		"goal": "criteria test",
		"acceptance_criteria": [
			{"description": "all tests pass", "validator": {"type": "test_pass"}}
		],
		"steps": [{"id": "s1", "goal": "build", "owner_agent": "op", "validator": {"type": "build_pass"}}]
	}`
	res, _ := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json": planJSON, "session_key": "s1", "original_request": "test",
	})
	runID := res.(map[string]interface{})["run_id"].(string)

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID, Type: EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})

	result, err := tm["run_propose_step_result"].call(execCtx, map[string]interface{}{
		"run_id": runID, "step_id": "s1", "result": "done",
	})
	require.NoError(t, err)
	m := result.(map[string]interface{})
	assert.Equal(t, "verified", m["status"])
	assert.Equal(t, "failed", m["run_status"])
	assert.NotNil(t, m["unmet_criteria"])

	snap, _ := store.GetRunSnapshot(ctx, runID)
	assert.Equal(t, RunStatusFailed, snap.Status)
}

func TestProposeResult_OrchestratorApproval_Flow(t *testing.T) {
	ctx := orchestratorCtx()
	execCtx := executionCtx()
	store := NewMemoryStore()
	pev := NewPEVEngine(store, DefaultValidators())
	tm := toolMap(store, pev)

	planJSON := `{
		"goal": "approval flow",
		"acceptance_criteria": [],
		"steps": [{"id": "s1", "goal": "review", "owner_agent": "op", "validator": {"type": "orchestrator_approval"}}]
	}`
	res, _ := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json": planJSON, "session_key": "s1", "original_request": "test",
	})
	runID := res.(map[string]interface{})["run_id"].(string)

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID, Type: EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})

	// Propose -> verification_failed (awaiting orchestrator approval).
	result, err := tm["run_propose_step_result"].call(execCtx, map[string]interface{}{
		"run_id": runID, "step_id": "s1", "result": "work done",
	})
	require.NoError(t, err)
	pm := result.(map[string]interface{})
	assert.Equal(t, "verification_failed", pm["status"])
	assert.Contains(t, pm["failure_reason"], "awaiting orchestrator approval")
}

func TestProposeResult_InfraError(t *testing.T) {
	execCtx := executionCtx()
	ctx := orchestratorCtx()
	store := NewMemoryStore()
	// No validators registered -> unknown type = infra error.
	pev := NewPEVEngine(store, map[ValidatorType]Validator{})
	tm := toolMap(store, pev)

	planJSON := `{
		"goal": "infra error",
		"acceptance_criteria": [],
		"steps": [{"id": "s1", "goal": "build", "owner_agent": "op", "validator": {"type": "build_pass"}}]
	}`
	res, _ := tm["run_create"].call(ctx, map[string]interface{}{
		"plan_json": planJSON, "session_key": "s1", "original_request": "test",
	})
	runID := res.(map[string]interface{})["run_id"].(string)

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID, Type: EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})

	_, err := tm["run_propose_step_result"].call(execCtx, map[string]interface{}{
		"run_id": runID, "step_id": "s1", "result": "done",
	})
	require.Error(t, err) // non-nil Go error for infra failure
	assert.Contains(t, err.Error(), "no validator registered")
}
