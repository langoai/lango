package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWorkflowToolsWave50_RunFromFileAppliesDefaultDelivery(t *testing.T) {
	t.Parallel()

	state := newMemoryRunStore()
	runner := &mockAgentRunner{result: "file-result"}
	sender := &recordingChannelSender{}
	engine := NewEngine(runner, state, sender, 1, time.Minute, zap.NewNop().Sugar())
	t.Cleanup(func() {
		require.NoError(t, engine.Shutdown(context.Background()))
	})

	workflowPath := filepath.Join(t.TempDir(), "from-file.flow.yaml")
	require.NoError(t, os.WriteFile(
		workflowPath,
		[]byte("name: from-file\nsteps:\n  - id: first\n    agent: executor\n    prompt: hello from file\n"),
		0o644,
	))
	tool := findWorkflowTool(t, BuildTools(engine, t.TempDir(), []string{"ops"}), "workflow_run")

	got, err := tool.Handler(context.Background(), map[string]interface{}{"file_path": workflowPath})
	require.NoError(t, err)
	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "running", payload["status"])
	assert.Equal(t, "run-1", payload["run_id"])
	assert.Contains(t, payload["message"], "Workflow 'from-file' started")

	state.waitForCompletion(t)
	sender.mu.Lock()
	defer sender.mu.Unlock()
	require.Len(t, sender.messages["ops"], 1)
	assert.Contains(t, sender.messages["ops"][0], "Workflow 'from-file' completed.")
	assert.Contains(t, sender.messages["ops"][0], "file-result")
}

func TestWorkflowToolsWave50_StatusListAndCancelHandlersReturnStructuredPayloads(t *testing.T) {
	t.Parallel()

	state := newMemoryRunStore()
	state.runs["run-1"] = &RunStatus{
		RunID:          "run-1",
		WorkflowName:   "stored",
		Status:         "running",
		TotalSteps:     2,
		CompletedSteps: 1,
		StepStatuses: []StepStatus{
			{StepID: "first", Agent: "operator", Status: "completed"},
			{StepID: "second", Agent: "executor", Status: "running"},
		},
	}
	state.stepResults["run-1"] = map[string]string{"first": "ok"}
	engine := NewEngine(nil, state, nil, 1, time.Minute, zap.NewNop().Sugar())
	cancelled := make(chan struct{})
	engine.mu.Lock()
	engine.cancels["run-1"] = func() { close(cancelled) }
	engine.mu.Unlock()
	tools := BuildTools(engine, t.TempDir(), nil)

	statusPayload, err := findWorkflowTool(t, tools, "workflow_status").Handler(
		context.Background(),
		map[string]interface{}{"run_id": "run-1"},
	)
	require.NoError(t, err)
	status, ok := statusPayload.(*RunStatus)
	require.True(t, ok)
	assert.Equal(t, "stored", status.WorkflowName)
	assert.Equal(t, "running", status.Status)
	assert.Equal(t, 1, status.CompletedSteps)
	require.Len(t, status.StepStatuses, 2)
	assert.Equal(t, "second", status.StepStatuses[1].StepID)

	listPayload, err := findWorkflowTool(t, tools, "workflow_list").Handler(
		context.Background(),
		map[string]interface{}{"limit": 5},
	)
	require.NoError(t, err)
	listMap, ok := listPayload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, listMap["count"])
	runs, ok := listMap["runs"].([]RunStatus)
	require.True(t, ok)
	require.Len(t, runs, 1)
	assert.Equal(t, "run-1", runs[0].RunID)

	cancelPayload, err := findWorkflowTool(t, tools, "workflow_cancel").Handler(
		context.Background(),
		map[string]interface{}{"run_id": "run-1"},
	)
	require.NoError(t, err)
	cancelMap, ok := cancelPayload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, map[string]interface{}{"status": "cancelled", "run_id": "run-1"}, cancelMap)
	select {
	case <-cancelled:
	default:
		t.Fatal("cancel handler did not invoke the registered cancel function")
	}
	statusAfterCancel, err := state.GetRunStatus(context.Background(), "run-1")
	require.NoError(t, err)
	assert.Equal(t, "cancelled", statusAfterCancel.Status)
}

func TestWorkflowToolsWave50_ListAndCancelPropagateEngineErrors(t *testing.T) {
	t.Parallel()

	state := newMemoryRunStore()
	state.listRunsErr = assert.AnError
	engine := NewEngine(nil, state, nil, 1, time.Minute, zap.NewNop().Sugar())
	tools := BuildTools(engine, t.TempDir(), nil)

	got, err := findWorkflowTool(t, tools, "workflow_list").Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "list workflow runs")

	got, err = findWorkflowTool(t, tools, "workflow_cancel").Handler(
		context.Background(),
		map[string]interface{}{"run_id": "missing"},
	)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, `cancel workflow: run "missing" not found or not running`)
}

func TestWorkflowToolsWave50_SaveWritesValidatedWorkflowFile(t *testing.T) {
	t.Parallel()

	engine := NewEngine(nil, newMemoryRunStore(), nil, 1, time.Minute, zap.NewNop().Sugar())
	stateDir := t.TempDir()
	tool := findWorkflowTool(t, BuildTools(engine, stateDir, nil), "workflow_save")
	yamlContent := "name: saved-flow\nsteps:\n  - id: first\n    agent: executor\n    prompt: persist me\n"

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"name":         "saved-flow",
		"yaml_content": yamlContent,
	})
	require.NoError(t, err)
	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "saved", payload["status"])
	assert.Equal(t, "saved-flow", payload["name"])
	filePath, ok := payload["file_path"].(string)
	require.True(t, ok)
	assert.Equal(t, filepath.Join(stateDir, "saved-flow.flow.yaml"), filePath)
	assert.Contains(t, payload["message"], "Workflow 'saved-flow' saved")

	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, yamlContent, string(content))
}

func TestWorkflowToolsWave50_SaveRejectsInvalidWorkflowDefinitions(t *testing.T) {
	t.Parallel()

	engine := NewEngine(nil, newMemoryRunStore(), nil, 1, time.Minute, zap.NewNop().Sugar())
	tool := findWorkflowTool(t, BuildTools(engine, t.TempDir(), nil), "workflow_save")

	testCases := []struct {
		name        string
		yamlContent string
		wantErr     string
	}{
		{
			name:        "parse error",
			yamlContent: "{{invalid yaml",
			wantErr:     "parse workflow YAML",
		},
		{
			name:        "validation error",
			yamlContent: "name: invalid\nsteps: []\n",
			wantErr:     "validate workflow",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tool.Handler(context.Background(), map[string]interface{}{
				"name":         "bad-flow",
				"yaml_content": tc.yamlContent,
			})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}
