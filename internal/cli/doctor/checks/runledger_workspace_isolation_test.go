package checks

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestRunLedgerWorkspaceIsolationCheck_Disabled(t *testing.T) {
	t.Parallel()
	cfg := config.DefaultConfig()
	cfg.RunLedger.WorkspaceIsolation = false

	c := &RunLedgerWorkspaceIsolationCheck{}
	r := c.Run(context.Background(), cfg)
	assert.Equal(t, StatusSkip, r.Status)
	assert.Contains(t, r.Message, "not enabled")
}

func TestRunLedgerWorkspaceIsolationCheck_NilConfig(t *testing.T) {
	t.Parallel()
	c := &RunLedgerWorkspaceIsolationCheck{}
	r := c.Run(context.Background(), nil)
	assert.Equal(t, StatusSkip, r.Status)
}

func TestRunLedgerWorkspaceIsolationCheck_Enabled(t *testing.T) {
	t.Parallel()
	cfg := config.DefaultConfig()
	cfg.RunLedger.WorkspaceIsolation = true

	c := &RunLedgerWorkspaceIsolationCheck{}
	r := c.Run(context.Background(), cfg)
	assert.NotEqual(t, StatusFail, r.Status)
}

func TestRunLedgerWorkspaceIsolationCheck_Name(t *testing.T) {
	t.Parallel()
	c := &RunLedgerWorkspaceIsolationCheck{}
	assert.Equal(t, "RunLedger Workspace Isolation", c.Name())
}
