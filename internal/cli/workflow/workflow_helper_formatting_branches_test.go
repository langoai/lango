package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

func TestWorkflowHelperFormattingBranches(t *testing.T) {
	assert.Equal(t, "workflow:Nightly Report", scheduledWorkflowJobName("Nightly Report"))

	prompt := scheduledWorkflowPrompt("/tmp/report.flow.yaml")
	assert.Contains(t, prompt, "workflow_run")
	assert.Contains(t, prompt, `/tmp/report.flow.yaml`)

	assert.Equal(t, "short", shortID("short"))
	assert.Equal(t, "12345678", shortID("123456789"))
	assert.Equal(t, "abcdef...", truncate("abcdefghij", 6))

	assert.Equal(t, "-", formatTime(time.Time{}))
	ts := time.Date(2026, 5, 15, 9, 8, 7, 0, time.UTC)
	assert.Equal(t, "2026-05-15 09:08:07", formatTime(ts))
}

func TestWorkflowRunLedgerListSkipsWhenAuthoritativeReadDisabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.AuthoritativeRead = false

	runs, handled, err := maybeListRunsFromLedger(&bootstrap.Result{Config: cfg}, 5)

	require.NoError(t, err)
	assert.False(t, handled)
	assert.Nil(t, runs)
}

func TestWorkflowRunLedgerListReportsUnavailableStorage(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.AuthoritativeRead = true

	runs, handled, err := maybeListRunsFromLedger(&bootstrap.Result{Config: cfg}, 5)

	require.Error(t, err)
	assert.True(t, handled)
	assert.Nil(t, runs)
	assert.Contains(t, err.Error(), "workflow runledger storage unavailable")
}

func TestWorkflowRunLedgerStatusSkipsWhenRunLedgerDisabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = false
	cfg.RunLedger.AuthoritativeRead = true

	status, handled, err := maybeStatusFromLedger(&bootstrap.Result{Config: cfg}, "run-1")

	require.NoError(t, err)
	assert.False(t, handled)
	assert.Nil(t, status)
}

func TestWorkflowRunLedgerStatusReportsUnavailableStorage(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.AuthoritativeRead = true

	status, handled, err := maybeStatusFromLedger(&bootstrap.Result{Config: cfg}, "run-1")

	require.Error(t, err)
	assert.True(t, handled)
	assert.Nil(t, status)
	assert.Contains(t, err.Error(), "workflow runledger storage unavailable")
}

func TestWorkflowStoresReturnNilWithoutBootstrapStorage(t *testing.T) {
	assert.Nil(t, workflowRunStore(nil, nil))
	assert.Nil(t, workflowRunLedgerStore(nil))
	assert.Nil(t, workflowRunStore(&bootstrap.Result{}, nil))
	assert.Nil(t, workflowRunLedgerStore(&bootstrap.Result{}))
}
