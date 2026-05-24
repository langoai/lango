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

func TestTeamRuntimeGuidance_UsesConcreteTeamTools(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		wantContains string
		wantNot      string
	}{
		{
			name: "list guidance",
			cmd: newTeamListCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         nil,
			wantContains: "Start the server with 'lango serve' and inspect/form teams via team_list, team_form, and team_form_with_budget.",
			wantNot:      "runtime integrations and agent tools",
		},
		{
			name: "status guidance",
			cmd: newTeamStatusCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         []string{"team-123"},
			wantContains: "Use the running server plus the team_status tool for live inspection.",
			wantNot:      "team runtime or agent tools",
		},
		{
			name: "disband guidance",
			cmd: newTeamDisbandCmd(func() (*bootstrap.Result, error) {
				return &bootstrap.Result{Config: cfg}, nil
			}),
			args:         []string{"team-123"},
			wantContains: "Use the running server plus the team_disband tool to disband a live team.",
			wantNot:      "team runtime or agent tools",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out, err := executeP2PCmd(t, tc.cmd, tc.args...)
			assert.NoError(t, err)
			assert.Contains(t, out, tc.wantContains)
			assert.NotContains(t, out, tc.wantNot)
		})
	}
}

func TestTeamCmd_HelpUsesConcreteToolSurfaceWording(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true

	cmd := newTeamCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	out, err := executeP2PCmd(t, cmd, "--help")
	require.NoError(t, err)

	assert.Contains(t, out, "team_form")
	assert.Contains(t, out, "team_form_with_budget")
	assert.Contains(t, out, "team_status")
	assert.Contains(t, out, "team_list")
	assert.Contains(t, out, "team_disband")
	assert.NotContains(t, out, "agent/tool-backed control paths")
}

func TestTeamListCmd_WritesJSONToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true

	cmd := newTeamListCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	out, err := executeP2PCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded []any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Empty(t, decoded)
}

func TestTeamStatusCmd_WritesJSONToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true

	cmd := newTeamStatusCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	out, err := executeP2PCmd(t, cmd, "team-123", "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "team not found (teams are runtime-only)", decoded["error"])
}
