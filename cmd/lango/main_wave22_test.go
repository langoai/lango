package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/types"
)

func TestWave22RunChatValidModeReachesAppBuilderBeforeTUI(t *testing.T) {
	restoreWave19TUIProfile(t)
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
	boot := &bootstrap.Result{Config: cfg, ProfileName: "wave22"}
	chatBootLoaderFn = func() (*bootstrap.Result, error) { return boot, nil }
	chatLoggingInitFn = func(logging.LogConfig) error { return nil }
	chatLoggingSyncFn = func() error { return nil }
	chatStartupErrWriter = &bytes.Buffer{}

	builderCalled := false
	chatAppBuilderFn = func(got *bootstrap.Result) (*app.App, error) {
		builderCalled = true
		assert.Same(t, boot, got)
		return nil, errors.New("stop before chat TUI")
	}

	err := runChat("research")

	require.Error(t, err)
	assert.True(t, builderCalled)
	assert.Contains(t, err.Error(), "create application: stop before chat TUI")
}

func TestWave22RunCockpitDefaultUsesLocalChatModeBeforeTUI(t *testing.T) {
	restoreWave19TUIProfile(t)
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
	boot := &bootstrap.Result{Config: cfg, ProfileName: "wave22"}
	withChannels = false
	cockpitBootLoaderFn = func() (*bootstrap.Result, error) { return boot, nil }
	cockpitLoggingInitFn = func(logging.LogConfig) error { return nil }
	cockpitLoggingSyncFn = func() error { return nil }
	cockpitStartupErrWriter = &bytes.Buffer{}

	builderCalled := false
	cockpitAppBuilderFn = func(got *bootstrap.Result, mode app.AppOption) (*app.App, error) {
		builderCalled = true
		assert.Same(t, boot, got)
		assert.Equal(t, app.AppModeLocalChat, wave19AppMode(t, mode))
		return nil, errors.New("stop before cockpit TUI")
	}

	err := runCockpit("")

	require.Error(t, err)
	assert.True(t, builderCalled)
	assert.Contains(t, err.Error(), "create application: stop before cockpit TUI")
}

func TestWave22RunWorkbenchPropagatesBootstrapAndLoggingErrors(t *testing.T) {
	restoreWave19TUIProfile(t)
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

	workbenchBootLoaderFn = func() (*bootstrap.Result, error) {
		return nil, errors.New("workbench boot refused")
	}
	workbenchAppBuilderFn = func(*bootstrap.Result) (*app.App, error) {
		t.Fatal("app builder must not run after bootstrap failure")
		return nil, nil
	}
	err := runWorkbench("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap: workbench boot refused")

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	workbenchBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "wave22"}, nil
	}
	workbenchLoggingInitFn = func(logging.LogConfig) error {
		return errors.New("workbench log refused")
	}
	workbenchLoggingSyncFn = func() error { return nil }
	workbenchStartupErrWriter = &bytes.Buffer{}
	err = runWorkbench("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "init logging: workbench log refused")
}

func TestWave22RunWorkbenchValidModeReachesAppBuilderBeforeTUI(t *testing.T) {
	restoreWave19TUIProfile(t)
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
		"debug": {Name: "debug"},
	}
	boot := &bootstrap.Result{Config: cfg, ProfileName: "wave22"}
	workbenchBootLoaderFn = func() (*bootstrap.Result, error) { return boot, nil }
	workbenchLoggingInitFn = func(logging.LogConfig) error { return nil }
	workbenchLoggingSyncFn = func() error { return nil }
	workbenchStartupErrWriter = &bytes.Buffer{}

	builderCalled := false
	workbenchAppBuilderFn = func(got *bootstrap.Result) (*app.App, error) {
		builderCalled = true
		assert.Same(t, boot, got)
		return nil, errors.New("stop before workbench TUI")
	}

	err := runWorkbench("debug")

	require.Error(t, err)
	assert.True(t, builderCalled)
	assert.Contains(t, err.Error(), "create application: stop before workbench TUI")
}

func TestWave22CockpitCommandWithChannelsFlagRoutesModeAndSetsGlobal(t *testing.T) {
	origInteractive := isInteractiveFn
	origRunner := runCockpitFn
	origWithChannels := withChannels
	t.Cleanup(func() {
		isInteractiveFn = origInteractive
		runCockpitFn = origRunner
		withChannels = origWithChannels
	})

	isInteractiveFn = func() bool { return true }
	withChannels = false
	var gotMode string
	runCockpitFn = func(mode string) error {
		gotMode = mode
		return nil
	}

	cmd := cockpitCmd()
	cmd.SetArgs([]string{"--mode", "debug", "--with-channels"})

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Equal(t, "debug", gotMode)
	assert.True(t, withChannels)
}

func TestWave22RegisterCockpitPagesStatusPageUsesFeatureStatuses(t *testing.T) {
	t.Parallel()

	statuses := app.NewStatusCollector()
	statuses.Add(&types.FeatureStatus{
		Name:    "Wave22 Feature",
		Enabled: false,
		Reason:  "not configured",
	})
	application := &app.App{
		Store:           stubCockpitSessionStore{},
		FeatureStatuses: statuses,
	}
	model := cockpit.New(cockpit.Deps{})
	chatModel := model.ChatModel()
	require.NotNil(t, chatModel)

	registerCockpitPages(
		model,
		application,
		config.DefaultConfig(),
		"wave22",
		nil,
		cockpit.Deps{},
		chatModel,
	)

	statusPage := model.Pages()[cockpit.PageStatus]
	require.NotNil(t, statusPage)
	_ = statusPage.Activate()

	view := statusPage.View()
	assert.Contains(t, view, "Wave22 Feature")
	assert.Contains(t, view, "not configured")
}

func TestWave22ConfigCommandValueSubcommandsAreQuietOnErrors(t *testing.T) {
	cmd := configCmd()

	for _, name := range []string{"get", "set"} {
		sub := findChildCommand(cmd.Commands(), name)
		require.NotNil(t, sub, "missing config %s subcommand", name)
		assert.True(t, sub.SilenceUsage, "config %s should not print usage on runtime errors", name)
		assert.True(t, sub.SilenceErrors, "config %s should let callers format errors", name)
	}
}
