package workflow

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	workflowpkg "github.com/langoai/lango/internal/workflow"
)

func TestWorkflowRun_WithScheduleReportsNotImplementedRegistration(t *testing.T) {
	workflowPath := writeWorkflowFixture(t, `name: Weekly Report
steps:
  - id: first
    agent: operator
    prompt: hello
`)

	cmd := newRunCmd(func() (*bootstrap.Result, error) {
		t.Fatal("boot loader should not be called when schedule registration is not implemented")
		return nil, nil
	})
	output, err := executeWorkflowRunCmd(t, cmd, workflowPath, "--schedule", "0 8 * * MON")

	require.NoError(t, err)
	require.Contains(t, output, "Workflow has a schedule.")
	require.Contains(t, output, "CLI schedule registration is not implemented yet.")
	require.Contains(t, output, "Use `lango cron add` or the runtime automation tools to schedule this workflow.")
	require.NotContains(t, output, "/api/workflow/register")
}

func TestWorkflowRun_WithoutScheduleReportsServerUnavailableOnCommandWriter(t *testing.T) {
	workflowPath := writeWorkflowFixture(t, `name: Daily Report
steps:
  - id: first
    agent: operator
    prompt: hello
`)

	cmd := newRunCmd(func() (*bootstrap.Result, error) {
		return nil, assert.AnError
	})
	output, err := executeWorkflowRunCmd(t, cmd, workflowPath)

	require.NoError(t, err)
	require.Contains(t, output, "Workflow validated successfully.")
	require.Contains(t, output, "(Server not available for direct execution)")
}

func TestWorkflowRun_WithoutScheduleReportsEngineDisabledOnCommandWriter(t *testing.T) {
	workflowPath := writeWorkflowFixture(t, `name: Daily Report
steps:
  - id: first
    agent: operator
    prompt: hello
`)

	cmd := newRunCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.Workflow.Enabled = false
		return &bootstrap.Result{Config: cfg}, nil
	})
	output, err := executeWorkflowRunCmd(t, cmd, workflowPath)

	require.NoError(t, err)
	require.Contains(t, output, "Workflow validated successfully.")
	require.Contains(t, output, "(Workflow engine not enabled in config)")
}

func TestWorkflowRun_WithoutScheduleWritesCompletionToCommandWriter(t *testing.T) {
	workflowPath := writeWorkflowFixture(t, `name: Daily Report
steps:
  - id: fetch-data
    agent: operator
    prompt: hello
`)

	orig := executeWorkflowDirect
	executeWorkflowDirect = func(_ *bootstrap.Result, _ *workflowpkg.Workflow) (*workflowpkg.RunResult, error) {
		return &workflowpkg.RunResult{
			Status: "completed",
			StepResults: map[string]string{
				"fetch-data": "Retrieved 42 records from the database...",
			},
		}, nil
	}
	t.Cleanup(func() { executeWorkflowDirect = orig })

	cmd := newRunCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.Workflow.Enabled = true
		return &bootstrap.Result{Config: cfg}, nil
	})
	output, err := executeWorkflowRunCmd(t, cmd, workflowPath)

	require.NoError(t, err)
	require.Contains(t, output, "Workflow validated successfully.")
	require.Contains(t, output, "Executing workflow...")
	require.Contains(t, output, "Workflow completed: completed")
	require.Contains(t, output, "--- Step: fetch-data ---")
	require.Contains(t, output, "Retrieved 42 records from the database...")
}

func writeWorkflowFixture(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	workflowPath := filepath.Join(dir, "report.flow.yaml")
	require.NoError(t, os.WriteFile(workflowPath, []byte(body), 0o644))
	return workflowPath
}

func executeWorkflowRunCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}
