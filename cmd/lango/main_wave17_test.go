package main

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
)

func TestConfigCmdRegistersLegacyAndValueSubcommands(t *testing.T) {
	t.Parallel()

	cmd := configCmd()

	assert.Equal(t, "sys", cmd.GroupID)
	for _, name := range []string{"list", "create", "use", "delete", "import", "export", "validate", "get", "set", "keys"} {
		assert.NotNil(t, cmd.Commands(), "config command should expose subcommands")
		assert.NotNil(t, findChildCommand(cmd.Commands(), name), "missing config subcommand %q", name)
	}
}

func TestRunChat_PropagatesBootstrapAndStartupErrors(t *testing.T) {
	origBootLoader := chatBootLoaderFn
	origLoggingInit := chatLoggingInitFn
	origStartupWriter := chatStartupErrWriter
	t.Cleanup(func() {
		chatBootLoaderFn = origBootLoader
		chatLoggingInitFn = origLoggingInit
		chatStartupErrWriter = origStartupWriter
	})

	chatBootLoaderFn = func() (*bootstrap.Result, error) {
		return nil, errors.New("boot unavailable")
	}
	err := runChat("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap: boot unavailable")

	chatStartupErrWriter = &bytes.Buffer{}
	chatBootLoaderFn = func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.DataRoot = t.TempDir()
		return &bootstrap.Result{Config: cfg, ProfileName: "test"}, nil
	}
	chatLoggingInitFn = func(logging.LogConfig) error {
		return errors.New("log init failed")
	}
	err = runChat("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "init logging: log init failed")
}

func TestRunCockpitAndWorkbenchPropagateAppBuilderErrors(t *testing.T) {
	origCockpitBoot := cockpitBootLoaderFn
	origCockpitLoggingInit := cockpitLoggingInitFn
	origCockpitLoggingSync := cockpitLoggingSyncFn
	origCockpitWriter := cockpitStartupErrWriter
	origCockpitBuilder := cockpitAppBuilderFn
	origWorkbenchBoot := workbenchBootLoaderFn
	origWorkbenchLoggingInit := workbenchLoggingInitFn
	origWorkbenchLoggingSync := workbenchLoggingSyncFn
	origWorkbenchWriter := workbenchStartupErrWriter
	origWorkbenchBuilder := workbenchAppBuilderFn
	t.Cleanup(func() {
		cockpitBootLoaderFn = origCockpitBoot
		cockpitLoggingInitFn = origCockpitLoggingInit
		cockpitLoggingSyncFn = origCockpitLoggingSync
		cockpitStartupErrWriter = origCockpitWriter
		cockpitAppBuilderFn = origCockpitBuilder
		workbenchBootLoaderFn = origWorkbenchBoot
		workbenchLoggingInitFn = origWorkbenchLoggingInit
		workbenchLoggingSyncFn = origWorkbenchLoggingSync
		workbenchStartupErrWriter = origWorkbenchWriter
		workbenchAppBuilderFn = origWorkbenchBuilder
	})

	cockpitStartupErrWriter = &bytes.Buffer{}
	workbenchStartupErrWriter = &bytes.Buffer{}
	cockpitLoggingInitFn = func(logging.LogConfig) error { return nil }
	cockpitLoggingSyncFn = func() error { return nil }
	workbenchLoggingInitFn = func(logging.LogConfig) error { return nil }
	workbenchLoggingSyncFn = func() error { return nil }

	cockpitBootLoaderFn = func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.DataRoot = t.TempDir()
		return &bootstrap.Result{Config: cfg, ProfileName: "test"}, nil
	}
	cockpitAppBuilderFn = func(*bootstrap.Result, app.AppOption) (*app.App, error) {
		return nil, errors.New("cockpit app failed")
	}
	err := runCockpit("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create application: cockpit app failed")

	workbenchBootLoaderFn = func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.DataRoot = t.TempDir()
		return &bootstrap.Result{Config: cfg, ProfileName: "test"}, nil
	}
	workbenchAppBuilderFn = func(*bootstrap.Result) (*app.App, error) {
		return nil, errors.New("workbench app failed")
	}
	err = runWorkbench("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create application: workbench app failed")
}

func TestBgTaskActionerRetryTaskReportsMissingAndSubmitErrors(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()
	mgr := background.NewManager(nil, nil, 1, time.Minute, logger)
	actioner := &bgTaskActioner{mgr: mgr}

	err := actioner.RetryTask(context.Background(), "missing-task")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `retry task missing-task: task status: task "missing-task" not found`)

	runner := &blockingBackgroundRunner{started: make(chan struct{})}
	mgr = background.NewManager(runner, nil, 1, time.Minute, logger)
	actioner = &bgTaskActioner{mgr: mgr}
	taskID, err := mgr.Submit(context.Background(), "original prompt", background.Origin{
		Channel: "cli",
		Session: "sess-1",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		require.NoError(t, mgr.Shutdown(shutdownCtx))
	})
	require.Eventually(t, func() bool {
		select {
		case <-runner.started:
			return true
		default:
			return false
		}
	}, 2*time.Second, 10*time.Millisecond)

	err = actioner.RetryTask(context.Background(), taskID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "submit task: max concurrent tasks reached (1)")
}

func TestTaskElapsedHandlesPendingCompletedAndRunningSnapshots(t *testing.T) {
	t.Parallel()

	assert.Zero(t, taskElapsed(background.TaskSnapshot{}))

	started := time.Now().Add(-10 * time.Second)
	completed := started.Add(3 * time.Second)
	assert.Equal(t, 3*time.Second, taskElapsed(background.TaskSnapshot{
		StartedAt:   started,
		CompletedAt: completed,
	}))

	running := taskElapsed(background.TaskSnapshot{StartedAt: time.Now().Add(-20 * time.Millisecond)})
	assert.Positive(t, running)
}

func TestRegisterCockpitPagesWiresTaskPageWhenBackgroundManagerExists(t *testing.T) {
	t.Parallel()

	model := cockpit.New(cockpit.Deps{})
	chatModel := model.ChatModel()
	require.NotNil(t, chatModel)
	application := &app.App{
		Store:             stubCockpitSessionStore{},
		BackgroundManager: background.NewManager(nil, nil, 1, time.Minute, zap.NewNop().Sugar()),
	}

	registerCockpitPages(
		model,
		application,
		config.DefaultConfig(),
		"profile",
		nil,
		cockpit.Deps{},
		chatModel,
	)

	_, ok := model.Pages()[cockpit.PageTasks]
	assert.True(t, ok)
	assert.False(t, model.Sidebar().IsDisabled(cockpit.PageTasks.String()))
}

func findChildCommand(commands []*cobra.Command, name string) *cobra.Command {
	for _, cmd := range commands {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
