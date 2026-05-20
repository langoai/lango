package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
)

var runChatBuilderErrorRunsStartupCleanupWithoutTuiMainSeamMu sync.Mutex

func TestRunChatBuilderErrorRunsStartupCleanupWithoutTUI(t *testing.T) {
	runChatBuilderErrorRunsStartupCleanupWithoutTuiMainSeamMu.Lock()
	defer runChatBuilderErrorRunsStartupCleanupWithoutTuiMainSeamMu.Unlock()
	restorePrepareTuiStartupInitializesLoggingAndRedirectsStdlibLogTUIProfile(t)

	origBootLoader := chatBootLoaderFn
	origLoggingInit := chatLoggingInitFn
	origLoggingSync := chatLoggingSyncFn
	origWriter := chatStartupErrWriter
	origBuilder := chatAppBuilderFn
	t.Cleanup(func() {
		chatBootLoaderFn = origBootLoader
		chatLoggingInitFn = origLoggingInit
		chatLoggingSyncFn = origLoggingSync
		chatStartupErrWriter = origWriter
		chatAppBuilderFn = origBuilder
	})

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	boot := &bootstrap.Result{Config: cfg, ProfileName: "runChatBuilderErrorRunsStartupCleanupWithoutTui9"}
	chatBootLoaderFn = func() (*bootstrap.Result, error) { return boot, nil }
	chatLoggingInitFn = func(logging.LogConfig) error { return nil }
	syncCalls := 0
	chatLoggingSyncFn = func() error {
		syncCalls++
		return errors.New("sync errors are cleanup-only")
	}
	var startup bytes.Buffer
	chatStartupErrWriter = &startup
	chatAppBuilderFn = func(got *bootstrap.Result) (*app.App, error) {
		assert.Same(t, boot, got)
		return nil, errors.New("runChatBuilderErrorRunsStartupCleanupWithoutTui9 stop before chat TUI")
	}

	err := runChat("")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create application: runChatBuilderErrorRunsStartupCleanupWithoutTui9 stop before chat TUI")
	assert.Equal(t, 1, syncCalls)
	assert.Contains(t, startup.String(), tui.Banner())
	assert.Contains(t, startup.String(), "Initializing...")
}

func TestRunCockpitWithChannelsBuilderErrorRunsStartupCleanupWithoutTUI(t *testing.T) {
	runChatBuilderErrorRunsStartupCleanupWithoutTuiMainSeamMu.Lock()
	defer runChatBuilderErrorRunsStartupCleanupWithoutTuiMainSeamMu.Unlock()
	restorePrepareTuiStartupInitializesLoggingAndRedirectsStdlibLogTUIProfile(t)

	origBootLoader := cockpitBootLoaderFn
	origLoggingInit := cockpitLoggingInitFn
	origLoggingSync := cockpitLoggingSyncFn
	origWriter := cockpitStartupErrWriter
	origBuilder := cockpitAppBuilderFn
	origWithChannels := withChannels
	t.Cleanup(func() {
		cockpitBootLoaderFn = origBootLoader
		cockpitLoggingInitFn = origLoggingInit
		cockpitLoggingSyncFn = origLoggingSync
		cockpitStartupErrWriter = origWriter
		cockpitAppBuilderFn = origBuilder
		withChannels = origWithChannels
	})

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	boot := &bootstrap.Result{Config: cfg, ProfileName: "runChatBuilderErrorRunsStartupCleanupWithoutTui9"}
	withChannels = true
	cockpitBootLoaderFn = func() (*bootstrap.Result, error) { return boot, nil }
	cockpitLoggingInitFn = func(logging.LogConfig) error { return nil }
	syncCalls := 0
	cockpitLoggingSyncFn = func() error {
		syncCalls++
		return nil
	}
	var startup bytes.Buffer
	cockpitStartupErrWriter = &startup
	cockpitAppBuilderFn = func(got *bootstrap.Result, mode app.AppOption) (*app.App, error) {
		assert.Same(t, boot, got)
		assert.Equal(t, app.AppModeCockpit, prepareTuiStartupInitializesLoggingAndRedirectsStdlibLogAppMode(t, mode))
		return nil, errors.New("runChatBuilderErrorRunsStartupCleanupWithoutTui9 stop before cockpit TUI")
	}

	err := runCockpit("")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create application: runChatBuilderErrorRunsStartupCleanupWithoutTui9 stop before cockpit TUI")
	assert.Equal(t, 1, syncCalls)
	assert.Contains(t, startup.String(), "Initializing cockpit...")
}

func TestRunWorkbenchBuilderErrorRunsStartupCleanupWithoutTUI(t *testing.T) {
	runChatBuilderErrorRunsStartupCleanupWithoutTuiMainSeamMu.Lock()
	defer runChatBuilderErrorRunsStartupCleanupWithoutTuiMainSeamMu.Unlock()
	restorePrepareTuiStartupInitializesLoggingAndRedirectsStdlibLogTUIProfile(t)

	origBootLoader := workbenchBootLoaderFn
	origLoggingInit := workbenchLoggingInitFn
	origLoggingSync := workbenchLoggingSyncFn
	origWriter := workbenchStartupErrWriter
	origBuilder := workbenchAppBuilderFn
	t.Cleanup(func() {
		workbenchBootLoaderFn = origBootLoader
		workbenchLoggingInitFn = origLoggingInit
		workbenchLoggingSyncFn = origLoggingSync
		workbenchStartupErrWriter = origWriter
		workbenchAppBuilderFn = origBuilder
	})

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	boot := &bootstrap.Result{Config: cfg, ProfileName: "runChatBuilderErrorRunsStartupCleanupWithoutTui9"}
	workbenchBootLoaderFn = func() (*bootstrap.Result, error) { return boot, nil }
	workbenchLoggingInitFn = func(logging.LogConfig) error { return nil }
	syncCalls := 0
	workbenchLoggingSyncFn = func() error {
		syncCalls++
		return nil
	}
	workbenchStartupErrWriter = io.Discard
	workbenchAppBuilderFn = func(got *bootstrap.Result) (*app.App, error) {
		assert.Same(t, boot, got)
		return nil, errors.New("runChatBuilderErrorRunsStartupCleanupWithoutTui9 stop before workbench TUI")
	}

	err := runWorkbench("")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create application: runChatBuilderErrorRunsStartupCleanupWithoutTui9 stop before workbench TUI")
	assert.Equal(t, 1, syncCalls)
}

func TestConfigSetMissingEnvFailsBeforeBootstrap(t *testing.T) {
	const missingEnvName = "LANGO_CONFIG_MISSING_ENV"
	previous, hadPrevious := os.LookupEnv(missingEnvName)
	require.NoError(t, os.Unsetenv(missingEnvName))
	t.Cleanup(func() {
		if hadPrevious {
			require.NoError(t, os.Setenv(missingEnvName, previous))
			return
		}
		require.NoError(t, os.Unsetenv(missingEnvName))
	})

	cmd := configCmd()
	cmd.SetArgs([]string{"set", "providers.openai.apiKey", "--from-env", missingEnvName})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), `environment variable "`+missingEnvName+`" is not set`)
}

func TestStartupSummaryRendersEnabledChannelsAndMCPServerCount(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Channels.Telegram.Enabled = true
	cfg.Channels.Discord.Enabled = true
	cfg.Channels.Slack.Enabled = true
	cfg.MCP.Enabled = true
	cfg.MCP.Servers = map[string]config.MCPServerConfig{
		"alpha": {},
		"beta":  {},
	}

	out := startupSummary(cfg)

	assert.Contains(t, out, "telegram, discord, slack")
	assert.Contains(t, out, "2 server(s)")
}
