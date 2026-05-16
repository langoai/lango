package p2p

import (
	"encoding/json"
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestWorkspaceRuntimeGuidance_UsesServerBackedRuntimeWording(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		wantContains string
	}{
		{
			name: "create guidance",
			cmd: newWorkspaceCreateCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         []string{"research-project", "--goal", "shared context"},
			wantContains: "Start the server with 'lango serve' and use p2p_workspace_create.",
		},
		{
			name: "list guidance",
			cmd: newWorkspaceListCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         nil,
			wantContains: "Start the server with 'lango serve' and use p2p_workspace_list, p2p_workspace_create, or p2p_workspace_join.",
		},
		{
			name: "status guidance",
			cmd: newWorkspaceStatusCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         []string{"workspace-123"},
			wantContains: "Use the running server plus the p2p_workspace_status or p2p_workspace_read tools for inspection.",
		},
		{
			name: "join guidance",
			cmd: newWorkspaceJoinCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         []string{"workspace-123"},
			wantContains: "Use 'lango serve' and the server-backed runtime or p2p_workspace_join tool.",
		},
		{
			name: "leave guidance",
			cmd: newWorkspaceLeaveCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         []string{"workspace-123"},
			wantContains: "Use 'lango serve' and the server-backed runtime or p2p_workspace_leave tool.",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out, err := executeP2PCmd(t, tc.cmd, tc.args...)
			assert.NoError(t, err)
			assert.Contains(t, out, tc.wantContains)
			assert.NotContains(t, out, "runtime API")
			assert.NotContains(t, out, "agent tools")
		})
	}
}

func TestWorkspaceCmd_HelpUsesConcreteToolSurfaceWording(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true

	cmd := newWorkspaceCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	out, err := executeP2PCmd(t, cmd, "--help")
	require.NoError(t, err)

	assert.Contains(t, out, "p2p_workspace_create")
	assert.Contains(t, out, "p2p_workspace_join")
	assert.Contains(t, out, "p2p_workspace_leave")
	assert.Contains(t, out, "p2p_workspace_list")
	assert.Contains(t, out, "p2p_workspace_status")
	assert.Contains(t, out, "p2p_workspace_read")
	assert.NotContains(t, out, "agent/tool-backed flows")
}

func TestWorkspaceCreateCmd_WritesJSONToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true

	cmd := newWorkspaceCreateCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	out, err := executeP2PCmd(t, cmd, "research-project", "--goal", "shared context", "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "research-project", decoded["name"])
	assert.Equal(t, "shared context", decoded["goal"])
}

func TestWorkspaceListCmd_WritesJSONToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true

	cmd := newWorkspaceListCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	out, err := executeP2PCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded []any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Empty(t, decoded)
}

func TestWorkspaceStatusCmd_WritesJSONToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true

	cmd := newWorkspaceStatusCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	out, err := executeP2PCmd(t, cmd, "workspace-123", "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "workspace not found (workspaces are runtime-only)", decoded["error"])
}
