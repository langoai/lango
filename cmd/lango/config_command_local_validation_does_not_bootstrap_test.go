package main

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
)

func TestConfigCommandLocalValidationDoesNotBootstrap(t *testing.T) {
	for _, tt := range []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "get rejects unsupported output before config load",
			args:    []string{"get", "agent.provider", "--output", "yaml"},
			wantErr: `unknown output format "yaml"`,
		},
		{
			name:    "set rejects empty from-env before bootstrap",
			args:    []string{"set", "agent.provider", "--from-env", " "},
			wantErr: "--from-env requires an environment variable name",
		},
		{
			name:    "set rejects from-env with explicit value before bootstrap",
			args:    []string{"set", "agent.provider", "openai", "--from-env", "OPENAI_API_KEY"},
			wantErr: "--from-env cannot be combined with a value argument",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := configCmd()
			cmd.SetArgs(tt.args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err := cmd.Execute()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestRunChatEmptyModeBypassesModeValidationBeforeLogging(t *testing.T) {
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
	cfg.Modes = map[string]config.SessionMode{
		"research": {Name: "research"},
	}
	chatBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "configCommandLocalValidationDoesNotBootstrap8"}, nil
	}
	loggingCalled := false
	chatLoggingInitFn = func(logging.LogConfig) error {
		loggingCalled = true
		return errors.New("configCommandLocalValidationDoesNotBootstrap8 chat logging refused")
	}
	chatLoggingSyncFn = func() error { return nil }
	chatStartupErrWriter = &bytes.Buffer{}
	chatAppBuilderFn = func(*bootstrap.Result) (*app.App, error) {
		t.Fatal("app builder must not run after chat logging failure")
		return nil, nil
	}

	err := runChat("")

	require.Error(t, err)
	assert.True(t, loggingCalled)
	assert.Contains(t, err.Error(), "init logging: configCommandLocalValidationDoesNotBootstrap8 chat logging refused")
}

func TestRunCockpitInvalidModeStopsBeforeStartup(t *testing.T) {
	restorePrepareTuiStartupInitializesLoggingAndRedirectsStdlibLogTUIProfile(t)
	origBootLoader := cockpitBootLoaderFn
	origLoggingInit := cockpitLoggingInitFn
	origLoggingSync := cockpitLoggingSyncFn
	origWriter := cockpitStartupErrWriter
	origBuilder := cockpitAppBuilderFn
	t.Cleanup(func() {
		cockpitBootLoaderFn = origBootLoader
		cockpitLoggingInitFn = origLoggingInit
		cockpitLoggingSyncFn = origLoggingSync
		cockpitStartupErrWriter = origWriter
		cockpitAppBuilderFn = origBuilder
	})

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	cfg.Modes = map[string]config.SessionMode{
		"debug": {Name: "debug"},
	}
	cockpitBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "configCommandLocalValidationDoesNotBootstrap8"}, nil
	}
	cockpitLoggingInitFn = func(logging.LogConfig) error {
		t.Fatal("logging must not initialize after invalid cockpit mode")
		return nil
	}
	cockpitLoggingSyncFn = func() error { return nil }
	cockpitStartupErrWriter = io.Discard
	cockpitAppBuilderFn = func(*bootstrap.Result, app.AppOption) (*app.App, error) {
		t.Fatal("app builder must not run after invalid cockpit mode")
		return nil, nil
	}

	err := runCockpit("missing-mode")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown mode "missing-mode"`)
}

func TestRunWorkbenchInvalidModeStopsBeforeStartup(t *testing.T) {
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
	cfg.Modes = map[string]config.SessionMode{
		"review": {Name: "review"},
	}
	workbenchBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "configCommandLocalValidationDoesNotBootstrap8"}, nil
	}
	workbenchLoggingInitFn = func(logging.LogConfig) error {
		t.Fatal("logging must not initialize after invalid workbench mode")
		return nil
	}
	workbenchLoggingSyncFn = func() error { return nil }
	workbenchStartupErrWriter = io.Discard
	workbenchAppBuilderFn = func(*bootstrap.Result) (*app.App, error) {
		t.Fatal("app builder must not run after invalid workbench mode")
		return nil, nil
	}

	err := runWorkbench("missing-mode")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown mode "missing-mode"`)
}

func TestRegisterCockpitPagesKeepsStubbedApplicationPagesAvailable(t *testing.T) {
	t.Parallel()

	model := cockpit.New(cockpit.Deps{})
	chatModel := model.ChatModel()
	require.NotNil(t, chatModel)

	registerCockpitPages(
		model,
		&app.App{Store: stubCockpitSessionStore{}},
		config.DefaultConfig(),
		"configCommandLocalValidationDoesNotBootstrap8",
		nil,
		cockpit.Deps{},
		chatModel,
	)

	for _, id := range []cockpit.PageID{
		cockpit.PageMissionControl,
		cockpit.PageTools,
		cockpit.PageStatus,
		cockpit.PageSettings,
		cockpit.PageSessions,
		cockpit.PageTasks,
		cockpit.PageDeadLetters,
		cockpit.PageApprovals,
	} {
		_, ok := model.Pages()[id]
		assert.True(t, ok, "expected %s page to be registered", id)
		assert.False(t, model.Sidebar().IsDisabled(id.String()), "expected %s page to be enabled", id)
	}

	sessionsPage := model.Pages()[cockpit.PageSessions]
	require.NotNil(t, sessionsPage)
	loadCmd := sessionsPage.Activate()
	require.NotNil(t, loadCmd)
	updated, followup := sessionsPage.Update(loadCmd())
	assert.Nil(t, followup)
	assert.Contains(t, updated.View(), "No sessions found.")
}
