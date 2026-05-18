package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/cliboot"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
)

func TestRunMain_RootCommandSuccessReturnsZeroAndPublishesVersion(t *testing.T) {
	origSandbox := isSandboxWorkerModeFn
	origBroker := isStorageBrokerModeFn
	origNewRoot := newRootCmdFn
	origStderr := mainStderr
	origVersion := Version
	origBuildTime := BuildTime
	origBootVersion := cliboot.Version
	t.Cleanup(func() {
		isSandboxWorkerModeFn = origSandbox
		isStorageBrokerModeFn = origBroker
		newRootCmdFn = origNewRoot
		mainStderr = origStderr
		Version = origVersion
		BuildTime = origBuildTime
		cliboot.Version = origBootVersion
		tui.SetVersionInfo(origVersion, origBuildTime)
	})

	Version = "wave9-test"
	BuildTime = "2026-05-18T00:00:00Z"
	isSandboxWorkerModeFn = func() bool { return false }
	isStorageBrokerModeFn = func() bool { return false }

	var executed bool
	newRootCmdFn = func() *cobra.Command {
		return &cobra.Command{
			Use: "lango",
			RunE: func(cmd *cobra.Command, args []string) error {
				executed = true
				return nil
			},
		}
	}
	var stderr bytes.Buffer
	mainStderr = &stderr

	code := runMain()

	assert.Equal(t, 0, code)
	assert.True(t, executed)
	assert.Empty(t, stderr.String())
	assert.Equal(t, "wave9-test", cliboot.Version)
}

func TestRunMain_RootCommandExecuteSuccessUsesRootCommand(t *testing.T) {
	origSandbox := isSandboxWorkerModeFn
	origBroker := isStorageBrokerModeFn
	origNewRoot := newRootCmdFn
	t.Cleanup(func() {
		isSandboxWorkerModeFn = origSandbox
		isStorageBrokerModeFn = origBroker
		newRootCmdFn = origNewRoot
	})

	isSandboxWorkerModeFn = func() bool { return false }
	isStorageBrokerModeFn = func() bool { return false }
	executed := false
	newRootCmdFn = func() *cobra.Command {
		return &cobra.Command{Use: "lango", RunE: func(*cobra.Command, []string) error {
			executed = true
			return nil
		}}
	}

	code := runMain()

	assert.Equal(t, 0, code)
	assert.True(t, executed)
}

func TestRunMain_RootCommandExecuteErrorReturnsOneForUnstructuredError(t *testing.T) {
	origSandbox := isSandboxWorkerModeFn
	origBroker := isStorageBrokerModeFn
	origNewRoot := newRootCmdFn
	origStderr := mainStderr
	t.Cleanup(func() {
		isSandboxWorkerModeFn = origSandbox
		isStorageBrokerModeFn = origBroker
		newRootCmdFn = origNewRoot
		mainStderr = origStderr
	})

	isSandboxWorkerModeFn = func() bool { return false }
	isStorageBrokerModeFn = func() bool { return false }
	newRootCmdFn = func() *cobra.Command {
		return &cobra.Command{
			Use:           "lango",
			SilenceUsage:  true,
			SilenceErrors: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				return errors.New("plain root failure")
			},
		}
	}
	var stderr bytes.Buffer
	mainStderr = &stderr

	code := runMain()

	assert.Equal(t, 1, code)
	assert.Equal(t, "plain root failure\n", stderr.String())
}

func TestConfigCmdConstructsProfileAndValueSubcommands(t *testing.T) {
	cmd := configCmd()

	assert.Equal(t, "config", cmd.Name())
	assert.Equal(t, "sys", cmd.GroupID)

	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}

	for _, want := range []string{"list", "create", "use", "delete", "import", "export", "validate", "get", "set", "keys"} {
		assert.True(t, names[want], "expected config subcommand %q", want)
	}
}

func TestMCPServerCountReportsOnlyWhenMCPEnabledWithServers(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{
			name: "disabled omits detail",
			cfg: &config.Config{
				MCP: config.MCPConfig{
					Enabled: false,
					Servers: map[string]config.MCPServerConfig{"fs": {}},
				},
			},
			want: "",
		},
		{
			name: "enabled without servers omits detail",
			cfg: &config.Config{
				MCP: config.MCPConfig{Enabled: true},
			},
			want: "",
		},
		{
			name: "enabled with servers reports count",
			cfg: &config.Config{
				MCP: config.MCPConfig{
					Enabled: true,
					Servers: map[string]config.MCPServerConfig{
						"browser": {Transport: "stdio"},
						"memory":  {Transport: "http"},
					},
				},
			},
			want: "2 server(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, mcpServerCount(tt.cfg))
		})
	}
}

func TestStartupSummaryRendersEnabledChannelAndMCPDetails(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Server.HTTPEnabled = true
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = 18789
	cfg.Channels.Telegram.Enabled = true
	cfg.Channels.Discord.Enabled = true
	cfg.Channels.Slack.Enabled = true
	cfg.Knowledge.Enabled = true
	cfg.MCP.Enabled = true
	cfg.MCP.Servers = map[string]config.MCPServerConfig{
		"browser": {Transport: "stdio"},
		"memory":  {Transport: "http"},
	}

	out := startupSummary(cfg)

	require.Contains(t, out, "Features:")
	assert.Contains(t, out, "Gateway")
	assert.Contains(t, out, "http://127.0.0.1:18789")
	assert.Contains(t, out, "Channels")
	assert.Contains(t, out, "telegram, discord, slack")
	assert.Contains(t, out, "Knowledge")
	assert.Contains(t, out, "MCP")
	assert.Contains(t, out, "2 server(s)")
}
