package workflow

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
	workflowpkg "github.com/langoai/lango/internal/workflow"
)

func TestWorkflowRun_WithScheduleRegistersCronBackedWorkflow(t *testing.T) {
	workflowPath := writeWorkflowFixture(t, `name: Weekly Report
deliver_to:
  - slack
steps:
  - id: first
    agent: operator
    prompt: hello
`)

	client := testutil.TestEntClient(t)
	cmd := newRunCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config: config.DefaultConfig(),
			Storage: storage.NewFacade(nil, nil, storage.WithCronFactory(func() cron.Store {
				return cron.NewEntStore(client)
			})),
		}, nil
	})
	output, err := executeWorkflowRunCmd(t, cmd, workflowPath, "--schedule", "0 8 * * MON")

	require.NoError(t, err)
	require.Contains(t, output, "Workflow: Weekly Report")
	require.Contains(t, output, "Schedule: 0 8 * * MON")
	require.Contains(t, output, "Scheduled workflow registered as cron job")
	require.NotContains(t, output, "/api/workflow/register")
	require.NotContains(t, output, "not implemented")

	store := cron.NewEntStore(client)
	jobs, err := store.List(context.Background())
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	require.Equal(t, "workflow:Weekly Report", jobs[0].Name)
	require.Equal(t, "cron", jobs[0].ScheduleType)
	require.Equal(t, "0 8 * * MON", jobs[0].Schedule)
	require.Equal(t, []string{"slack"}, jobs[0].DeliverTo)
	require.True(t, jobs[0].Enabled)
	require.Contains(t, jobs[0].Prompt, "workflow_run")
	require.Contains(t, jobs[0].Prompt, workflowPath)
}

func TestWorkflowRun_WithScheduleReturnsBootstrapError(t *testing.T) {
	workflowPath := writeWorkflowFixture(t, `name: Weekly Report
steps:
  - id: first
    agent: operator
    prompt: hello
`)

	cmd := newRunCmd(func() (*bootstrap.Result, error) {
		return nil, assert.AnError
	})
	output, err := executeWorkflowRunCmd(t, cmd, workflowPath, "--schedule", "0 8 * * MON")

	require.Error(t, err)
	require.Contains(t, err.Error(), "bootstrap")
	require.Contains(t, output, "Workflow: Weekly Report")
	require.NotContains(t, output, "Scheduled workflow registered")
}

func TestWorkflowRun_WithScheduleReturnsCronStorageUnavailable(t *testing.T) {
	workflowPath := writeWorkflowFixture(t, `name: Weekly Report
steps:
  - id: first
    agent: operator
    prompt: hello
`)

	cmd := newRunCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: config.DefaultConfig()}, nil
	})
	output, err := executeWorkflowRunCmd(t, cmd, workflowPath, "--schedule", "0 8 * * MON")

	require.Error(t, err)
	require.Contains(t, err.Error(), "cron storage is not configured")
	require.Contains(t, output, "Workflow: Weekly Report")
	require.NotContains(t, output, "Scheduled workflow registered")
}

func TestWorkflowRun_WithScheduleRejectsInvalidCronSchedule(t *testing.T) {
	workflowPath := writeWorkflowFixture(t, `name: Weekly Report
steps:
  - id: first
    agent: operator
    prompt: hello
`)

	client := testutil.TestEntClient(t)
	cmd := newRunCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config: config.DefaultConfig(),
			Storage: storage.NewFacade(nil, nil, storage.WithCronFactory(func() cron.Store {
				return cron.NewEntStore(client)
			})),
		}, nil
	})
	output, err := executeWorkflowRunCmd(t, cmd, workflowPath, "--schedule", "not-cron")

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid workflow schedule")
	require.Contains(t, output, "Workflow: Weekly Report")
	require.NotContains(t, output, "Scheduled workflow registered")

	jobs, listErr := cron.NewEntStore(client).List(context.Background())
	require.NoError(t, listErr)
	require.Empty(t, jobs)
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
