package workflow

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/storage"
	workflowpkg "github.com/langoai/lango/internal/workflow"
)

func TestWorkflowRunCommandValidationBranchesAvoidBootstrap(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		workflowYML string
		wantErr     string
	}{
		{
			name:    "missing argument",
			args:    nil,
			wantErr: "accepts 1 arg(s), received 0",
		},
		{
			name:        "invalid yaml",
			workflowYML: "name: [\n",
			wantErr:     "parse workflow",
		},
		{
			name: "invalid workflow dependency",
			workflowYML: `
name: bad-dependency
steps:
  - id: collect
    agent: operator
    prompt: collect
    depends_on:
      - missing
`,
			wantErr: "validate workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calledBootstrap := false
			cmd := newRunCmd(func() (*bootstrap.Result, error) {
				calledBootstrap = true
				return nil, assert.AnError
			})
			args := append([]string(nil), tt.args...)
			if tt.workflowYML != "" {
				args = append(args, writeWorkflowFile(t, tt.workflowYML))
			}

			out, err := executeWorkflowCommand(t, cmd, args...)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.False(t, calledBootstrap)
			assert.Contains(t, out, "Usage:")
		})
	}
}

func TestWorkflowRunScheduleRegistrationFailureBranches(t *testing.T) {
	workflowPath := writeWorkflowFile(t, `
name: scheduled-report
steps:
  - id: collect
    agent: operator
    prompt: collect
`)

	t.Run("cron storage unavailable", func(t *testing.T) {
		cmd := newRunCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Config: config.DefaultConfig()}, nil
		})

		out, err := executeWorkflowCommand(t, cmd, workflowPath, "--schedule", "0 8 * * MON")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "workflow schedule registration unavailable")
		assert.Contains(t, out, "Workflow has a schedule.")
		assert.NotContains(t, out, "Scheduled workflow registered")
	})

	t.Run("invalid cron rejected before storage create", func(t *testing.T) {
		store := &workflowRunCommandValidationBranchesAvoidBootstrapCronStore{}
		cmd := newRunCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{
				Config: config.DefaultConfig(),
				Storage: storage.NewFacade(nil, nil, storage.WithCronFactory(func() cron.Store {
					return store
				})),
			}, nil
		})

		out, err := executeWorkflowCommand(t, cmd, workflowPath, "--schedule", "not-cron")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid workflow schedule")
		assert.Zero(t, store.createCalls)
		assert.Contains(t, out, "Schedule: not-cron")
		assert.NotContains(t, out, "Scheduled workflow registered")
	})
}

func TestWorkflowRunDirectExecutionBranches(t *testing.T) {
	workflowPath := writeWorkflowFile(t, `
name: direct-report
steps:
  - id: collect
    agent: operator
    prompt: collect
`)

	t.Run("bootstrap unavailable prints guidance without failing command", func(t *testing.T) {
		cmd := newRunCmd(func() (*bootstrap.Result, error) {
			return nil, assert.AnError
		})

		out, err := executeWorkflowCommand(t, cmd, workflowPath)

		require.NoError(t, err)
		assert.Contains(t, out, "Workflow validated successfully.")
		assert.Contains(t, out, "(Server not available for direct execution)")
	})

	t.Run("engine disabled prints config guidance", func(t *testing.T) {
		workflowRootCommandConstructsExpectedSubcommandsWorkflowSeamMu.Lock()
		original := executeWorkflowDirect
		executeWorkflowDirect = func(_ *bootstrap.Result, _ *workflowpkg.Workflow) (*workflowpkg.RunResult, error) {
			return nil, ErrWorkflowDisabled
		}
		t.Cleanup(func() {
			executeWorkflowDirect = original
			workflowRootCommandConstructsExpectedSubcommandsWorkflowSeamMu.Unlock()
		})

		cmd := newRunCmd(func() (*bootstrap.Result, error) {
			cfg := config.DefaultConfig()
			cfg.Workflow.Enabled = false
			return &bootstrap.Result{Config: cfg}, nil
		})

		out, err := executeWorkflowCommand(t, cmd, workflowPath)

		require.NoError(t, err)
		assert.Contains(t, out, "(Workflow engine not enabled in config)")
		assert.NotContains(t, out, "Executing workflow...")
	})

	t.Run("execution error is returned with context", func(t *testing.T) {
		workflowRootCommandConstructsExpectedSubcommandsWorkflowSeamMu.Lock()
		original := executeWorkflowDirect
		executeWorkflowDirect = func(_ *bootstrap.Result, _ *workflowpkg.Workflow) (*workflowpkg.RunResult, error) {
			return nil, errors.New("executor failed")
		}
		t.Cleanup(func() {
			executeWorkflowDirect = original
			workflowRootCommandConstructsExpectedSubcommandsWorkflowSeamMu.Unlock()
		})

		cmd := newRunCmd(func() (*bootstrap.Result, error) {
			cfg := config.DefaultConfig()
			cfg.Workflow.Enabled = true
			return &bootstrap.Result{Config: cfg}, nil
		})

		out, err := executeWorkflowCommand(t, cmd, workflowPath)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "execute workflow: executor failed")
		assert.Contains(t, out, "Workflow validated successfully.")
		assert.NotContains(t, out, "Workflow completed:")
	})

	t.Run("successful execution prints status and step results", func(t *testing.T) {
		workflowRootCommandConstructsExpectedSubcommandsWorkflowSeamMu.Lock()
		original := executeWorkflowDirect
		executeWorkflowDirect = func(_ *bootstrap.Result, w *workflowpkg.Workflow) (*workflowpkg.RunResult, error) {
			assert.Equal(t, "direct-report", w.Name)
			return &workflowpkg.RunResult{
				Status: "completed",
				StepResults: map[string]string{
					"collect": "done",
				},
			}, nil
		}
		t.Cleanup(func() {
			executeWorkflowDirect = original
			workflowRootCommandConstructsExpectedSubcommandsWorkflowSeamMu.Unlock()
		})

		cmd := newRunCmd(func() (*bootstrap.Result, error) {
			cfg := config.DefaultConfig()
			cfg.Workflow.Enabled = true
			return &bootstrap.Result{Config: cfg}, nil
		})

		out, err := executeWorkflowCommand(t, cmd, workflowPath)

		require.NoError(t, err)
		assert.Contains(t, out, "Executing workflow...")
		assert.Contains(t, out, "Workflow completed: completed")
		assert.Contains(t, out, "--- Step: collect ---")
		assert.Contains(t, out, "done")
	})
}

func TestWorkflowReadAndCancelOutputBranches(t *testing.T) {
	t.Run("list disabled emits command error", func(t *testing.T) {
		cmd := newWorkflowListCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Config: config.DefaultConfig()}, nil
		})

		out, err := executeWorkflowCommand(t, cmd)

		require.ErrorIs(t, err, ErrWorkflowDisabled)
		assert.Contains(t, out, "Error: workflow engine is not enabled")
		assert.Contains(t, out, "Usage:")
	})

	t.Run("status disabled emits command error", func(t *testing.T) {
		cmd := newWorkflowStatusCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Config: config.DefaultConfig()}, nil
		})

		out, err := executeWorkflowCommand(t, cmd, "run-47")

		require.ErrorIs(t, err, ErrWorkflowDisabled)
		assert.Contains(t, out, "Error: workflow engine is not enabled")
		assert.Contains(t, out, "Usage:")
	})

	t.Run("history disabled emits command error", func(t *testing.T) {
		cmd := newWorkflowHistoryCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Config: config.DefaultConfig()}, nil
		})

		out, err := executeWorkflowCommand(t, cmd)

		require.ErrorIs(t, err, ErrWorkflowDisabled)
		assert.Contains(t, out, "Error: workflow engine is not enabled")
		assert.Contains(t, out, "Usage:")
	})

	t.Run("cancel writes seam success message", func(t *testing.T) {
		workflowRootCommandConstructsExpectedSubcommandsWorkflowSeamMu.Lock()
		original := cancelWorkflowRun
		cancelWorkflowRun = func(_ func() (*bootstrap.Result, error), runID string) (string, error) {
			return "cancelled " + runID, nil
		}
		t.Cleanup(func() {
			cancelWorkflowRun = original
			workflowRootCommandConstructsExpectedSubcommandsWorkflowSeamMu.Unlock()
		})

		cmd := newWorkflowCancelCmd(func() (*bootstrap.Result, error) {
			return nil, assert.AnError
		})

		out, err := executeWorkflowCommand(t, cmd, "run-47")

		require.NoError(t, err)
		assert.Equal(t, "cancelled run-47\n", out)
	})
}

type workflowRunCommandValidationBranchesAvoidBootstrapCronStore struct {
	createCalls int
}

func (s *workflowRunCommandValidationBranchesAvoidBootstrapCronStore) Create(context.Context, cron.Job) error {
	s.createCalls++
	return nil
}

func (s *workflowRunCommandValidationBranchesAvoidBootstrapCronStore) Get(context.Context, string) (*cron.Job, error) {
	return nil, errors.New("not implemented")
}

func (s *workflowRunCommandValidationBranchesAvoidBootstrapCronStore) GetByName(context.Context, string) (*cron.Job, error) {
	return nil, errors.New("not implemented")
}

func (s *workflowRunCommandValidationBranchesAvoidBootstrapCronStore) List(context.Context) ([]cron.Job, error) {
	return nil, errors.New("not implemented")
}

func (s *workflowRunCommandValidationBranchesAvoidBootstrapCronStore) ListEnabled(context.Context) ([]cron.Job, error) {
	return nil, errors.New("not implemented")
}

func (s *workflowRunCommandValidationBranchesAvoidBootstrapCronStore) Update(context.Context, cron.Job) error {
	return errors.New("not implemented")
}

func (s *workflowRunCommandValidationBranchesAvoidBootstrapCronStore) Upsert(context.Context, cron.Job) (*cron.Job, bool, error) {
	return nil, false, errors.New("not implemented")
}

func (s *workflowRunCommandValidationBranchesAvoidBootstrapCronStore) Delete(context.Context, string) error {
	return errors.New("not implemented")
}

func (s *workflowRunCommandValidationBranchesAvoidBootstrapCronStore) SaveHistory(context.Context, cron.HistoryEntry) error {
	return errors.New("not implemented")
}

func (s *workflowRunCommandValidationBranchesAvoidBootstrapCronStore) ListHistory(context.Context, string, int) ([]cron.HistoryEntry, error) {
	return nil, errors.New("not implemented")
}

func (s *workflowRunCommandValidationBranchesAvoidBootstrapCronStore) ListAllHistory(context.Context, int) ([]cron.HistoryEntry, error) {
	return nil, errors.New("not implemented")
}
