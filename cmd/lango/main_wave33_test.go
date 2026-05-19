package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestWave33ConfigCommandWiresInspectionAndMutationSubcommands(t *testing.T) {
	t.Parallel()

	cmd := configCmd()
	names := make(map[string]bool)
	for _, child := range cmd.Commands() {
		names[child.Name()] = true
	}

	for _, name := range []string{"get", "set", "keys", "list", "use", "delete", "export", "import", "validate"} {
		assert.True(t, names[name], "expected config subcommand %q", name)
	}
	assert.Equal(t, "sys", cmd.GroupID)
}

func TestWave33ValidateInitialModeRejectsUnknownMode(t *testing.T) {
	t.Parallel()

	err := validateInitialMode(config.DefaultConfig(), "wave33-missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown mode "wave33-missing"`)
	assert.Contains(t, err.Error(), "valid modes can be listed via /mode")
}

func TestWave33StartupSummaryRendersDisabledChannelAndMCPDetails(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Channels.Telegram.Enabled = false
	cfg.Channels.Discord.Enabled = false
	cfg.Channels.Slack.Enabled = false
	cfg.MCP.Enabled = true
	cfg.MCP.Servers = nil

	out := startupSummary(cfg)

	assert.Contains(t, out, "Channels")
	assert.NotContains(t, out, "telegram")
	assert.NotContains(t, out, "discord")
	assert.NotContains(t, out, "slack")
	assert.Contains(t, out, "MCP")
	assert.NotContains(t, out, "server(s)")
}
