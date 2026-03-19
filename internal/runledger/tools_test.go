package runledger

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// orchestratorCtx returns a context that identifies the caller as the orchestrator.
// In real usage, toolchain.AgentNameFromContext is used, but in tests we test
// the default (empty = orchestrator) behavior.
func orchestratorCtx() context.Context {
	return context.Background()
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

func TestRunCreate_EndToEnd(t *testing.T) {
	ctx := orchestratorCtx()
	store := NewMemoryStore()
	pev := NewPEVEngine(store, DefaultValidators())
	tools := BuildTools(store, pev)

	// Find run_create tool.
	var createTool, readTool, activeTool, noteTool, proposeTool *runCreateHelper
	for _, tool := range tools {
		switch tool.Name {
		case "run_create":
			createTool = &runCreateHelper{tool.Handler}
		case "run_read":
			readTool = &runCreateHelper{tool.Handler}
		case "run_active":
			activeTool = &runCreateHelper{tool.Handler}
		case "run_note":
			noteTool = &runCreateHelper{tool.Handler}
		case "run_propose_step_result":
			proposeTool = &runCreateHelper{tool.Handler}
		}
	}

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

	result, err := createTool.call(ctx, map[string]interface{}{
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
	readResult, err := readTool.call(ctx, map[string]interface{}{"run_id": runID})
	require.NoError(t, err)
	snap := readResult.(*RunSnapshot)
	assert.Equal(t, RunStatusRunning, snap.Status)
	assert.Len(t, snap.Steps, 2)

	// Get active step.
	activeResult, err := activeTool.call(ctx, map[string]interface{}{"run_id": runID})
	require.NoError(t, err)
	am := activeResult.(map[string]interface{})
	// s1 should be the next available step.
	assert.Equal(t, "next_available", am["status"])

	// Write a note.
	_, err = noteTool.call(ctx, map[string]interface{}{
		"run_id": runID,
		"key":    "scratch",
		"value":  "testing in progress",
	})
	require.NoError(t, err)

	// Read note back.
	noteResult, err := noteTool.call(ctx, map[string]interface{}{
		"run_id": runID,
		"key":    "scratch",
	})
	require.NoError(t, err)
	nm := noteResult.(map[string]interface{})
	assert.Equal(t, "testing in progress", nm["value"])

	// Propose step result for s1.
	_, err = proposeTool.call(ctx, map[string]interface{}{
		"run_id":  runID,
		"step_id": "s1",
		"result":  "code written successfully",
	})
	require.NoError(t, err)

	// Verify step status changed to verify_pending.
	snap2, err := store.GetRunSnapshot(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, StepStatusVerifyPending, snap2.Steps[0].Status)
}

func TestRunCreate_InvalidPlan(t *testing.T) {
	ctx := orchestratorCtx()
	store := NewMemoryStore()
	pev := NewPEVEngine(store, DefaultValidators())
	tools := BuildTools(store, pev)

	var createTool *runCreateHelper
	for _, tool := range tools {
		if tool.Name == "run_create" {
			createTool = &runCreateHelper{tool.Handler}
		}
	}

	// Malformed JSON.
	result, err := createTool.call(ctx, map[string]interface{}{
		"plan_json":        "not json",
		"session_key":      "s1",
		"original_request": "test",
	})
	require.NoError(t, err) // returns error as structured result, not Go error
	m := result.(map[string]interface{})
	assert.Equal(t, "parse_failed", m["error"])

	// Valid JSON but fails validation.
	result2, err := createTool.call(ctx, map[string]interface{}{
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
	tools := BuildTools(store, pev)

	toolMap := make(map[string]*runCreateHelper, len(tools))
	for _, tool := range tools {
		toolMap[tool.Name] = &runCreateHelper{tool.Handler}
	}

	// Create run.
	planJSON := `{
		"goal": "test retry",
		"acceptance_criteria": [],
		"steps": [{"id": "s1", "goal": "do", "owner_agent": "op", "validator": {"type": "build_pass"}}]
	}`
	res, _ := toolMap["run_create"].call(ctx, map[string]interface{}{
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
	res2, err := toolMap["run_apply_policy"].call(ctx, map[string]interface{}{
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
	store := NewMemoryStore()
	pev := NewPEVEngine(store, DefaultValidators())
	tools := BuildTools(store, pev)

	toolMap := make(map[string]*runCreateHelper, len(tools))
	for _, tool := range tools {
		toolMap[tool.Name] = &runCreateHelper{tool.Handler}
	}

	planJSON := `{
		"goal": "test approval",
		"acceptance_criteria": [],
		"steps": [{"id": "s1", "goal": "review", "owner_agent": "op", "validator": {"type": "orchestrator_approval"}}]
	}`
	res, _ := toolMap["run_create"].call(ctx, map[string]interface{}{
		"plan_json": planJSON, "session_key": "s1", "original_request": "test",
	})
	runID := res.(map[string]interface{})["run_id"].(string)

	// Start step and propose result.
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "s1"}),
	})
	_, _ = toolMap["run_propose_step_result"].call(ctx, map[string]interface{}{
		"run_id":  runID,
		"step_id": "s1",
		"result":  "work done",
	})

	// Approve the step.
	approveRes, err := toolMap["run_approve_step"].call(ctx, map[string]interface{}{
		"run_id":  runID,
		"step_id": "s1",
		"reason":  "looks good",
	})
	require.NoError(t, err)
	am := approveRes.(map[string]interface{})
	assert.Equal(t, "approved", am["status"])

	// Verify step completed.
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
