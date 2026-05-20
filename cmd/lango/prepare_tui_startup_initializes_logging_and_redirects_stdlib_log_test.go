package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
)

type prepareTuiStartupInitializesLoggingAndRedirectsStdlibLogRoundTripFunc func(*http.Request) (*http.Response, error)

func (f prepareTuiStartupInitializesLoggingAndRedirectsStdlibLogRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func restorePrepareTuiStartupInitializesLoggingAndRedirectsStdlibLogTUIProfile(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { tui.SetProfile("default") })
}

func prepareTuiStartupInitializesLoggingAndRedirectsStdlibLogAppMode(t *testing.T, opt app.AppOption) app.AppMode {
	t.Helper()

	fn := reflect.ValueOf(opt)
	require.Equal(t, reflect.Func, fn.Kind())
	require.Equal(t, 1, fn.Type().NumIn())

	argType := fn.Type().In(0)
	require.Equal(t, reflect.Pointer, argType.Kind())

	holder := reflect.New(argType.Elem())
	fn.Call([]reflect.Value{holder})

	mode := holder.Elem().FieldByName("mode")
	require.True(t, mode.IsValid())
	return app.AppMode(mode.Int())
}

func TestPrepareTUIStartupInitializesLoggingAndRedirectsStdlibLog(t *testing.T) {
	restorePrepareTuiStartupInitializesLoggingAndRedirectsStdlibLogTUIProfile(t)
	origLogWriter := log.Writer()
	t.Cleanup(func() { log.SetOutput(origLogWriter) })

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	cfg.Logging.Level = "debug"
	cfg.Logging.Format = "console"

	var gotLogConfig logging.LogConfig
	syncCalled := false
	var notice bytes.Buffer

	cleanup, err := prepareTUIStartup(
		cfg,
		"prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9-profile",
		"prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9.log",
		func(logCfg logging.LogConfig) error {
			gotLogConfig = logCfg
			return nil
		},
		func() error {
			syncCalled = true
			return nil
		},
		&notice,
		"  Initializing prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9...",
	)
	require.NoError(t, err)

	log.Print("prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9 stdlib log")
	cleanup()

	logPath := filepath.Join(cfg.DataRoot, "prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9.log")
	assert.Equal(t, logging.LogConfig{
		Level:      "debug",
		Format:     "console",
		OutputPath: logPath,
	}, gotLogConfig)
	assert.True(t, syncCalled)
	assert.Contains(t, notice.String(), logPath)
	assert.Contains(t, notice.String(), "Initializing prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9...")

	logBytes, readErr := os.ReadFile(logPath)
	require.NoError(t, readErr)
	assert.Contains(t, string(logBytes), "prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9 stdlib log")
}

func TestPrepareTUIStartupReturnsInitLoggingErrorWithoutNotice(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	var notice bytes.Buffer

	cleanup, err := prepareTUIStartup(
		cfg,
		"prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9-profile",
		"prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9.log",
		func(logging.LogConfig) error { return errors.New("logger refused config") },
		func() error {
			t.Fatal("sync must not run when logging init fails")
			return nil
		},
		&notice,
		"  Initializing prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9...",
	)

	require.Error(t, err)
	assert.Nil(t, cleanup)
	assert.Contains(t, err.Error(), "init logging: logger refused config")
	assert.Empty(t, notice.String())
}

func TestHealthCmdUsesInjectedHTTPClientAndPort(t *testing.T) {
	origClientFn := newHealthHTTPClientFn
	t.Cleanup(func() { newHealthHTTPClientFn = origClientFn })

	var gotURL string
	newHealthHTTPClientFn = func() *http.Client {
		return &http.Client{Transport: prepareTuiStartupInitializesLoggingAndRedirectsStdlibLogRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(nil)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		})}
	}

	cmd := healthCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--port", "19019"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:19019/health", gotURL)
	assert.Equal(t, "ok\n", out.String())
}

func TestServeCmdPropagatesSetupErrorsBeforeStartingApplication(t *testing.T) {
	restorePrepareTuiStartupInitializesLoggingAndRedirectsStdlibLogTUIProfile(t)
	origBootLoader := serveBootLoaderFn
	origLoggingInit := serveLoggingInitFn
	origAppBuilder := serveAppBuilderFn
	t.Cleanup(func() {
		serveBootLoaderFn = origBootLoader
		serveLoggingInitFn = origLoggingInit
		serveAppBuilderFn = origAppBuilder
	})

	serveBootLoaderFn = func() (*bootstrap.Result, error) {
		return nil, errors.New("boot missing")
	}
	err := serveCmd().Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap: boot missing")

	cfg := config.DefaultConfig()
	serveBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9"}, nil
	}
	serveLoggingInitFn = func(logging.LogConfig) error {
		return errors.New("log path rejected")
	}
	err = serveCmd().Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "init logging: log path rejected")

	serveLoggingInitFn = func(logging.LogConfig) error { return nil }
	serveAppBuilderFn = func(*bootstrap.Result) (stoppableApplication, error) {
		return nil, errors.New("app builder stopped")
	}
	err = serveCmd().Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create application: app builder stopped")
}

func TestRunMainStorageBrokerModeShortCircuitsRootCommand(t *testing.T) {
	origSandbox := isSandboxWorkerModeFn
	origBroker := isStorageBrokerModeFn
	origRunBroker := runStorageBrokerServerFn
	origNewRoot := newRootCmdFn
	origStdin := mainStdin
	origStdout := mainStdout
	origStderr := mainStderr
	t.Cleanup(func() {
		isSandboxWorkerModeFn = origSandbox
		isStorageBrokerModeFn = origBroker
		runStorageBrokerServerFn = origRunBroker
		newRootCmdFn = origNewRoot
		mainStdin = origStdin
		mainStdout = origStdout
		mainStderr = origStderr
	})

	isSandboxWorkerModeFn = func() bool { return false }
	isStorageBrokerModeFn = func() bool { return true }
	mainStdin = bytes.NewBufferString("broker input")
	var stdout bytes.Buffer
	mainStdout = &stdout
	mainStderr = &bytes.Buffer{}
	runStorageBrokerServerFn = func(in io.Reader, out io.Writer) error {
		body, err := io.ReadAll(in)
		require.NoError(t, err)
		assert.Equal(t, "broker input", string(body))
		_, err = out.Write([]byte("broker ok"))
		return err
	}
	newRootCmdFn = func() *cobra.Command {
		t.Fatal("root command must not be constructed in storage broker mode")
		return nil
	}

	code := runMain()

	assert.Equal(t, 0, code)
	assert.Equal(t, "broker ok", stdout.String())
}

func TestRunCockpitWithChannelsFlagReachesCockpitAppModeSetup(t *testing.T) {
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
	withChannels = true
	cockpitStartupErrWriter = &bytes.Buffer{}
	cockpitBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9"}, nil
	}
	cockpitLoggingInitFn = func(logging.LogConfig) error { return nil }
	cockpitLoggingSyncFn = func() error { return nil }
	builderCalled := false
	cockpitAppBuilderFn = func(_ *bootstrap.Result, mode app.AppOption) (*app.App, error) {
		builderCalled = true
		assert.Equal(t, app.AppModeCockpit, prepareTuiStartupInitializesLoggingAndRedirectsStdlibLogAppMode(t, mode))
		return nil, errors.New("stop before TUI")
	}

	err := runCockpit("")

	require.Error(t, err)
	assert.True(t, builderCalled)
	assert.Contains(t, err.Error(), "create application: stop before TUI")
}

func TestValidateInitialModeAcceptsConfiguredMode(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Modes = map[string]config.SessionMode{
		"research": {Name: "research"},
	}

	require.NoError(t, validateInitialMode(cfg, "research"))
	require.NoError(t, validateInitialMode(cfg, ""))
}

func TestRunChatAppBuilderErrorSkipsApplicationStart(t *testing.T) {
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
		return &bootstrap.Result{Config: cfg, ProfileName: "prepareTuiStartupInitializesLoggingAndRedirectsStdlibLog9"}, nil
	}
	chatLoggingInitFn = func(logging.LogConfig) error { return nil }
	chatLoggingSyncFn = func() error { return nil }
	chatStartupErrWriter = &bytes.Buffer{}
	chatAppBuilderFn = func(*bootstrap.Result) (*app.App, error) {
		return nil, errors.New("chat app unavailable")
	}

	err := runChat("")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create application: chat app unavailable")
}

func TestWatchServeSignalsReturnsWhenSignalChannelCloses(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal)
	done := make(chan struct{})
	go func() {
		watchServeSignals(ctx, &fakeServeApp{}, zap.NewNop().Sugar(), sigChan, 10, cancel, func(int) {
			t.Fatal("closed signal channel must not force exit")
		})
		close(done)
	}()

	close(sigChan)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("watchServeSignals did not return")
	}
}
