package p2p

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

func TestGitFetchCmd_UsesTruthfulRuntimeGuidance(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true

	cmd := newGitFetchCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	out, err := executeP2PCmd(t, cmd, "workspace-123")
	assert.NoError(t, err)

	assert.Contains(t, out, "Fetch requires a running server.")
	assert.Contains(t, out, "Use 'lango serve' and the server-backed runtime plus provenance or workspace artifact tools for live exchange.")
	assert.NotContains(t, out, "p2p_git_fetch")
}

func TestGitRuntimeGuidance_UsesServerBackedRuntimeWording(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		wantContains string
	}{
		{
			name: "init guidance",
			cmd: newGitInitCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         []string{"workspace-123"},
			wantContains: "Use 'lango serve' and the server-backed runtime or p2p_git_init tool.",
		},
		{
			name: "log guidance",
			cmd: newGitLogCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         []string{"workspace-123"},
			wantContains: "Use the server-backed runtime or p2p_git_* tools for live repository inspection.",
		},
		{
			name: "diff guidance",
			cmd: newGitDiffCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         []string{"workspace-123", "abc123", "def456"},
			wantContains: "Use 'lango serve' and the server-backed runtime or p2p_git_diff tool.",
		},
		{
			name: "push guidance",
			cmd: newGitPushCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         []string{"workspace-123"},
			wantContains: "Use 'lango serve' and the server-backed runtime or p2p_git_push tool.",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out, err := executeP2PCmd(t, tc.cmd, tc.args...)
			assert.NoError(t, err)
			assert.Contains(t, out, tc.wantContains)
			assert.NotContains(t, out, "runtime API")
		})
	}
}

func TestGitLogCmd_WritesJSONToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true

	cmd := newGitLogCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	out, err := executeP2PCmd(t, cmd, "workspace-123", "--output", "json")
	assert.NoError(t, err)

	var decoded []any
	assert.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Empty(t, decoded)
}
