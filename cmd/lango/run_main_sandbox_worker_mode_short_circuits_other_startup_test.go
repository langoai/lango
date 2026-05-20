package main

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/cliexit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
)

func TestRunMainSandboxWorkerModeShortCircuitsOtherStartup(t *testing.T) {
	origSandbox := isSandboxWorkerModeFn
	origRunSandbox := runSandboxWorkerFn
	origBroker := isStorageBrokerModeFn
	origNewRoot := newRootCmdFn
	t.Cleanup(func() {
		isSandboxWorkerModeFn = origSandbox
		runSandboxWorkerFn = origRunSandbox
		isStorageBrokerModeFn = origBroker
		newRootCmdFn = origNewRoot
	})

	isSandboxWorkerModeFn = func() bool { return true }
	runSandboxWorkerFn = func() int { return 24 }
	isStorageBrokerModeFn = func() bool {
		t.Fatal("storage broker mode must not be checked after sandbox worker mode")
		return false
	}
	newRootCmdFn = func() *cobra.Command {
		t.Fatal("root command must not be constructed in sandbox worker mode")
		return nil
	}

	assert.Equal(t, 24, runMain())
}

func TestRunMainFormatsExitErrors(t *testing.T) {
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

	var stderr bytes.Buffer
	mainStderr = &stderr
	newRootCmdFn = func() *cobra.Command {
		return &cobra.Command{RunE: func(*cobra.Command, []string) error {
			return cliexit.New(7, errors.New("runMainSandboxWorkerModeShortCircuitsOtherStartup4 routed exit"))
		}}
	}
	assert.Equal(t, 7, runMain())
	assert.Contains(t, stderr.String(), "runMainSandboxWorkerModeShortCircuitsOtherStartup4 routed exit")

	stderr.Reset()
	newRootCmdFn = func() *cobra.Command {
		return &cobra.Command{RunE: func(*cobra.Command, []string) error {
			return cliexit.NewSilent(8)
		}}
	}
	assert.Equal(t, 8, runMain())
	assert.Empty(t, stderr.String())
}

func TestConfigCommandWiresProfileAndValueSubcommands(t *testing.T) {
	cmd := configCmd()

	assert.Equal(t, "sys", cmd.GroupID)
	for _, name := range []string{"list", "create", "use", "delete", "import", "export", "validate", "get", "set", "keys"} {
		assert.NotNil(t, findChildCommand(cmd.Commands(), name), "missing config %s subcommand", name)
	}
}

func TestRunChatInvalidModeStopsBeforeLoggingAndAppCreation(t *testing.T) {
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
	chatBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "runMainSandboxWorkerModeShortCircuitsOtherStartup4"}, nil
	}
	chatLoggingInitFn = func(logging.LogConfig) error {
		t.Fatal("logging must not initialize after invalid initial mode")
		return nil
	}
	chatLoggingSyncFn = func() error { return nil }
	chatStartupErrWriter = io.Discard
	chatAppBuilderFn = func(*bootstrap.Result) (*app.App, error) {
		t.Fatal("app builder must not run after invalid initial mode")
		return nil, nil
	}

	err := runChat("missing-mode")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown mode "missing-mode"`)
}

func TestRunCockpitLoggingErrorStopsBeforeAppBuilder(t *testing.T) {
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
	cockpitBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "runMainSandboxWorkerModeShortCircuitsOtherStartup4"}, nil
	}
	cockpitLoggingInitFn = func(logging.LogConfig) error {
		return errors.New("cockpit logging refused")
	}
	cockpitLoggingSyncFn = func() error { return nil }
	cockpitStartupErrWriter = io.Discard
	cockpitAppBuilderFn = func(*bootstrap.Result, app.AppOption) (*app.App, error) {
		t.Fatal("app builder must not run after cockpit logging failure")
		return nil, nil
	}

	err := runCockpit("")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "init logging: cockpit logging refused")
}

func TestRunWorkbenchBootstrapErrorStopsBeforeLogging(t *testing.T) {
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

	workbenchBootLoaderFn = func() (*bootstrap.Result, error) {
		return nil, errors.New("workbench bootstrap refused")
	}
	workbenchLoggingInitFn = func(logging.LogConfig) error {
		t.Fatal("logging must not initialize after bootstrap failure")
		return nil
	}
	workbenchLoggingSyncFn = func() error { return nil }
	workbenchStartupErrWriter = io.Discard
	workbenchAppBuilderFn = func(*bootstrap.Result) (*app.App, error) {
		t.Fatal("app builder must not run after bootstrap failure")
		return nil, nil
	}

	err := runWorkbench("")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap: workbench bootstrap refused")
}
