package workflow

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	workflowpkg "github.com/langoai/lango/internal/workflow"
)

var wave38WorkflowSeamMu sync.Mutex

func TestWave38WorkflowRootCommandConstructsExpectedSubcommands(t *testing.T) {
	cmd := NewWorkflowCmd(func() (*bootstrap.Result, error) {
		return nil, assert.AnError
	})

	assert.Equal(t, "workflow", cmd.Use)
	assert.Equal(t, "Manage workflow execution", cmd.Short)
	assert.Contains(t, cmd.Long, ".flow.yaml")

	var names []string
	for _, child := range cmd.Commands() {
		names = append(names, child.Name())
	}
	assert.ElementsMatch(t,
		[]string{"cancel", "history", "list", "run", "status", "validate"},
		names)
}

func TestWave38WorkflowRunRejectsInvalidWorkflowBeforeBootstrap(t *testing.T) {
	workflowPath := writeWorkflowFile(t, `
name: missing-steps
`)
	calledBootstrap := false
	cmd := newRunCmd(func() (*bootstrap.Result, error) {
		calledBootstrap = true
		return nil, assert.AnError
	})

	out, err := executeWorkflowCommand(t, cmd, workflowPath)

	require.Error(t, err)
	assert.False(t, calledBootstrap)
	assert.Contains(t, err.Error(), "parse workflow")
	assert.Contains(t, err.Error(), "validate workflow")
	assert.Contains(t, out, "Error: parse workflow")
	assert.Contains(t, out, "Usage:")
}

func TestWave38WorkflowValidateJSONReportsParseErrorToCommandWriter(t *testing.T) {
	workflowPath := writeWorkflowFile(t, "name: [\n")
	cmd := newValidateCmd()

	out, err := executeWorkflowCommand(t, cmd, workflowPath, "--output", "json")

	require.NoError(t, err)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, false, payload["valid"])
	assert.Equal(t, workflowPath, payload["file"])
	assert.Contains(t, payload["error"], "parse workflow YAML")
}

func TestWave38WorkflowRunFormatsExecutionErrorAndTruncatedStepOutput(t *testing.T) {
	workflowPath := writeWorkflowFile(t, `
name: Daily Report
steps:
  - id: collect
    agent: operator
    prompt: collect
`)
	longResult := strings.Repeat("x", 700)

	wave38WorkflowSeamMu.Lock()
	original := executeWorkflowDirect
	executeWorkflowDirect = func(_ *bootstrap.Result, _ *workflowpkg.Workflow) (*workflowpkg.RunResult, error) {
		return &workflowpkg.RunResult{
			Status: "failed",
			Error:  "step failed",
			StepResults: map[string]string{
				"collect": longResult,
			},
		}, nil
	}
	t.Cleanup(func() {
		executeWorkflowDirect = original
		wave38WorkflowSeamMu.Unlock()
	})

	cmd := newRunCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.Workflow.Enabled = true
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeWorkflowCommand(t, cmd, workflowPath)

	require.NoError(t, err)
	assert.Contains(t, out, "Workflow completed: failed")
	assert.Contains(t, out, "Error: step failed")
	assert.Contains(t, out, "--- Step: collect ---")
	assert.Contains(t, out, strings.Repeat("x", 500)+"...")
	assert.NotContains(t, out, longResult)
}

func TestWave38WorkflowCancelPropagatesErrorWithoutOutput(t *testing.T) {
	wave38WorkflowSeamMu.Lock()
	original := cancelWorkflowRun
	cancelWorkflowRun = func(_ func() (*bootstrap.Result, error), _ string) (string, error) {
		return "", assert.AnError
	}
	t.Cleanup(func() {
		cancelWorkflowRun = original
		wave38WorkflowSeamMu.Unlock()
	})

	cmd := newWorkflowCancelCmd(func() (*bootstrap.Result, error) {
		return nil, nil
	})

	out, err := executeWorkflowCommand(t, cmd, "run-38")

	require.ErrorIs(t, err, assert.AnError)
	assert.Contains(t, out, "Error: assert.AnError general error for testing")
	assert.Contains(t, out, "Usage:")
}

func TestWave38WorkflowReadCommandsReturnBootstrapErrors(t *testing.T) {
	bootLoader := func() (*bootstrap.Result, error) {
		return nil, assert.AnError
	}

	tests := []struct {
		name string
		cmd  func(t *testing.T) string
	}{
		{
			name: "list",
			cmd: func(t *testing.T) string {
				out, err := executeWorkflowCommand(t, newWorkflowListCmd(bootLoader))
				require.Error(t, err)
				assert.Contains(t, err.Error(), "bootstrap")
				return out
			},
		},
		{
			name: "status",
			cmd: func(t *testing.T) string {
				out, err := executeWorkflowCommand(t, newWorkflowStatusCmd(bootLoader), "run-38")
				require.Error(t, err)
				assert.Contains(t, err.Error(), "bootstrap")
				return out
			},
		},
		{
			name: "history",
			cmd: func(t *testing.T) string {
				out, err := executeWorkflowCommand(t, newWorkflowHistoryCmd(bootLoader))
				require.Error(t, err)
				assert.Contains(t, err.Error(), "bootstrap")
				return out
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tt.cmd(t)
			assert.Contains(t, out, "Error: bootstrap: assert.AnError general error for testing")
			assert.Contains(t, out, "Usage:")
		})
	}
}
