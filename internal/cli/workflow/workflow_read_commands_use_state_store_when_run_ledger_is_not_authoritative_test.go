package workflow

import (
	"context"
	"fmt"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/storage"
	workflowpkg "github.com/langoai/lango/internal/workflow"
)

func TestWorkflowReadCommandsUseStateStoreWhenRunLedgerIsNotAuthoritative(t *testing.T) {
	t.Run("list prints started column from workflow state store", func(t *testing.T) {
		bootLoader, runID := workflowReadCommandsUseStateStoreWhenRunLedgerIsNotAuthoritativeWorkflowBootLoader(t)
		out, err := executeWorkflowCommand(t, newWorkflowListCmd(bootLoader), "--limit", "3")

		require.NoError(t, err)
		assert.Contains(t, out, "ID")
		assert.Contains(t, out, "STARTED")
		assert.Contains(t, out, runID[:8])
		assert.Contains(t, out, "runChatUsesProgramSeamAndCleansUpSession1-state")
		assert.Contains(t, out, "completed")
		assert.Contains(t, out, "2/2")
	})

	t.Run("status prints state-store step errors with truncation", func(t *testing.T) {
		bootLoader, runID := workflowReadCommandsUseStateStoreWhenRunLedgerIsNotAuthoritativeWorkflowBootLoader(t)
		out, err := executeWorkflowCommand(t, newWorkflowStatusCmd(bootLoader), runID)

		require.NoError(t, err)
		assert.Contains(t, out, "Run ID:    "+runID)
		assert.Contains(t, out, "Workflow:  runChatUsesProgramSeamAndCleansUpSession1-state")
		assert.Contains(t, out, "Progress:  2/2 steps")
		assert.Contains(t, out, "collect")
		assert.Contains(t, out, "operator")
		assert.Contains(t, out, "completed")
		assert.Contains(t, out, "review")
		assert.Contains(t, out, "validator")
		assert.Contains(t, out, "validator "+strings.Repeat("x", 30)+"...")
	})

	t.Run("history prints compact table from workflow state store", func(t *testing.T) {
		bootLoader, _ := workflowReadCommandsUseStateStoreWhenRunLedgerIsNotAuthoritativeWorkflowBootLoader(t)
		out, err := executeWorkflowCommand(t, newWorkflowHistoryCmd(bootLoader), "-n", "1")

		require.NoError(t, err)
		assert.Contains(t, out, "ID")
		assert.Contains(t, out, "WORKFLOW")
		assert.Contains(t, out, "STEPS")
		assert.Contains(t, out, "runChatUsesProgramSeamAndCleansUpSession1-state")
		assert.Contains(t, out, "2/2")
		assert.NotContains(t, out, "STARTED")
	})
}

func TestWorkflowReadCommandsWrapStateStoreErrors(t *testing.T) {
	client := enttest.Open(t, "sqlite3", workflowReadCommandsUseStateStoreWhenRunLedgerIsNotAuthoritativeSQLiteDSN(t, "closed"))
	cfg := config.DefaultConfig()
	cfg.Workflow.Enabled = true
	facade := storage.NewFacade(nil, nil, storage.WithEntClient(client))
	require.NoError(t, client.Close())
	bootLoader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, Storage: facade}, nil
	}

	t.Run("list wraps storage error", func(t *testing.T) {
		out, err := executeWorkflowCommand(t, newWorkflowListCmd(bootLoader))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "list runs:")
		assert.Contains(t, out, "Error: list runs:")
	})

	t.Run("status wraps storage error", func(t *testing.T) {
		out, err := executeWorkflowCommand(t, newWorkflowStatusCmd(bootLoader), "bad-run")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "get status:")
		assert.Contains(t, out, "Error: get status:")
	})

	t.Run("history wraps storage error", func(t *testing.T) {
		out, err := executeWorkflowCommand(t, newWorkflowHistoryCmd(bootLoader))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "list runs:")
		assert.Contains(t, out, "Error: list runs:")
	})
}

func TestInitEngineBranches(t *testing.T) {
	disabled := config.DefaultConfig()
	disabled.Workflow.Enabled = false
	assert.Nil(t, initEngine(&bootstrap.Result{Config: disabled}))

	enabledWithoutStorage := config.DefaultConfig()
	enabledWithoutStorage.Workflow.Enabled = true
	assert.Nil(t, initEngine(&bootstrap.Result{Config: enabledWithoutStorage}))

	client := enttest.Open(t, "sqlite3", workflowReadCommandsUseStateStoreWhenRunLedgerIsNotAuthoritativeSQLiteDSN(t, "engine"))
	t.Cleanup(func() { client.Close() })
	enabled := config.DefaultConfig()
	enabled.Workflow.Enabled = true
	engine := initEngine(&bootstrap.Result{
		Config:  enabled,
		Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
	})

	require.NotNil(t, engine)
}

func workflowReadCommandsUseStateStoreWhenRunLedgerIsNotAuthoritativeWorkflowBootLoader(t *testing.T) (func() (*bootstrap.Result, error), string) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", workflowReadCommandsUseStateStoreWhenRunLedgerIsNotAuthoritativeSQLiteDSN(t, "state"))
	t.Cleanup(func() { client.Close() })

	state := workflowpkg.NewStateStore(client, nil)
	ctx := context.Background()
	wf := &workflowpkg.Workflow{
		Name: "runChatUsesProgramSeamAndCleansUpSession1-state",
		Steps: []workflowpkg.Step{
			{ID: "collect", Agent: "operator", Prompt: "collect"},
			{ID: "review", Agent: "validator", Prompt: "review"},
		},
	}
	runID, err := state.CreateRun(ctx, wf)
	require.NoError(t, err)
	require.NoError(t, state.CreateStepRun(ctx, runID, wf.Steps[0], "collect"))
	require.NoError(t, state.UpdateStepStatus(ctx, runID, "collect", "completed", "done", ""))
	require.NoError(t, state.CreateStepRun(ctx, runID, wf.Steps[1], "review"))
	require.NoError(t, state.UpdateStepStatus(
		ctx,
		runID,
		"review",
		"failed",
		"",
		"validator "+strings.Repeat("x", 80),
	))
	require.NoError(t, state.CompleteRun(ctx, runID, "completed", ""))

	cfg := config.DefaultConfig()
	cfg.Workflow.Enabled = true
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.AuthoritativeRead = false
	facade := storage.NewFacade(nil, nil, storage.WithEntClient(client))
	return func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, Storage: facade}, nil
	}, runID
}

func workflowReadCommandsUseStateStoreWhenRunLedgerIsNotAuthoritativeSQLiteDSN(t *testing.T, suffix string) string {
	t.Helper()
	name := strings.NewReplacer("/", "-", " ", "-", "_", "-").Replace(t.Name())
	return fmt.Sprintf("file:%s-%s?mode=memory&_fk=1", name, suffix)
}
