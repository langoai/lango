package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/session"
)

type wave46SessionStore struct {
	stubCockpitSessionStore

	createErr error
	endKeys   []string
}

func (s *wave46SessionStore) Create(*session.Session) error {
	return s.createErr
}

func (s *wave46SessionStore) End(key string) error {
	s.endKeys = append(s.endKeys, key)
	return nil
}

func TestWave46RunChatStartErrorStopsBeforeSessionAndTUI(t *testing.T) {
	restoreWave19TUIProfile(t)
	origBootLoader := chatBootLoaderFn
	origLoggingInit := chatLoggingInitFn
	origLoggingSync := chatLoggingSyncFn
	origWriter := chatStartupErrWriter
	origBuilder := chatAppBuilderFn
	origStart := startAppFn
	origStop := stopAppFn
	t.Cleanup(func() {
		chatBootLoaderFn = origBootLoader
		chatLoggingInitFn = origLoggingInit
		chatLoggingSyncFn = origLoggingSync
		chatStartupErrWriter = origWriter
		chatAppBuilderFn = origBuilder
		startAppFn = origStart
		stopAppFn = origStop
	})

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	boot := &bootstrap.Result{Config: cfg, ProfileName: "wave46"}
	store := &wave46SessionStore{}
	startErr := errors.New("wave46 app start refused")
	chatBootLoaderFn = func() (*bootstrap.Result, error) { return boot, nil }
	chatLoggingInitFn = func(logging.LogConfig) error { return nil }
	chatLoggingSyncFn = func() error { return nil }
	chatStartupErrWriter = io.Discard
	chatAppBuilderFn = func(got *bootstrap.Result) (*app.App, error) {
		assert.Same(t, boot, got)
		return &app.App{Store: store}, nil
	}
	startAppFn = func(*app.App, context.Context) error { return startErr }
	stopAppFn = func(*app.App, context.Context) error {
		t.Fatal("stop must not run when app start fails")
		return nil
	}

	err := runChat("")

	require.ErrorIs(t, err, startErr)
	assert.Contains(t, err.Error(), "start application: wave46 app start refused")
	assert.Empty(t, store.endKeys)
}

func TestWave46RunCockpitInitialSessionCreateErrorStopsBeforeTUIAndStopsApp(t *testing.T) {
	restoreWave19TUIProfile(t)
	origBootLoader := cockpitBootLoaderFn
	origLoggingInit := cockpitLoggingInitFn
	origLoggingSync := cockpitLoggingSyncFn
	origWriter := cockpitStartupErrWriter
	origBuilder := cockpitAppBuilderFn
	origStart := startAppFn
	origStop := stopAppFn
	origWithChannels := withChannels
	t.Cleanup(func() {
		cockpitBootLoaderFn = origBootLoader
		cockpitLoggingInitFn = origLoggingInit
		cockpitLoggingSyncFn = origLoggingSync
		cockpitStartupErrWriter = origWriter
		cockpitAppBuilderFn = origBuilder
		startAppFn = origStart
		stopAppFn = origStop
		withChannels = origWithChannels
	})

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	cfg.Modes = map[string]config.SessionMode{"ops": {Name: "ops"}}
	boot := &bootstrap.Result{Config: cfg, ProfileName: "wave46"}
	store := &wave46SessionStore{createErr: errors.New("wave46 create refused")}
	var startup bytes.Buffer
	stopCalls := 0
	withChannels = false
	cockpitBootLoaderFn = func() (*bootstrap.Result, error) { return boot, nil }
	cockpitLoggingInitFn = func(logging.LogConfig) error { return nil }
	cockpitLoggingSyncFn = func() error { return nil }
	cockpitStartupErrWriter = &startup
	cockpitAppBuilderFn = func(got *bootstrap.Result, mode app.AppOption) (*app.App, error) {
		assert.Same(t, boot, got)
		assert.Equal(t, app.AppModeLocalChat, wave19AppMode(t, mode))
		return &app.App{Store: store, EventBus: eventbus.New()}, nil
	}
	startAppFn = func(*app.App, context.Context) error { return nil }
	stopAppFn = func(*app.App, context.Context) error {
		stopCalls++
		return nil
	}

	err := runCockpit("ops")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create initial session: wave46 create refused")
	assert.Equal(t, 1, stopCalls)
	assert.Contains(t, startup.String(), "Initializing cockpit...")
	require.Len(t, store.endKeys, 1)
	assert.Contains(t, store.endKeys[0], "cockpit-")
}

func TestWave46RunWorkbenchInitialSessionCreateErrorStopsBeforeTUIAndStopsApp(t *testing.T) {
	restoreWave19TUIProfile(t)
	origBootLoader := workbenchBootLoaderFn
	origLoggingInit := workbenchLoggingInitFn
	origLoggingSync := workbenchLoggingSyncFn
	origWriter := workbenchStartupErrWriter
	origBuilder := workbenchAppBuilderFn
	origStart := startAppFn
	origStop := stopAppFn
	t.Cleanup(func() {
		workbenchBootLoaderFn = origBootLoader
		workbenchLoggingInitFn = origLoggingInit
		workbenchLoggingSyncFn = origLoggingSync
		workbenchStartupErrWriter = origWriter
		workbenchAppBuilderFn = origBuilder
		startAppFn = origStart
		stopAppFn = origStop
	})

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	cfg.Modes = map[string]config.SessionMode{"review": {Name: "review"}}
	boot := &bootstrap.Result{Config: cfg, ProfileName: "wave46"}
	store := &wave46SessionStore{createErr: errors.New("wave46 workbench create refused")}
	stopCalls := 0
	workbenchBootLoaderFn = func() (*bootstrap.Result, error) { return boot, nil }
	workbenchLoggingInitFn = func(logging.LogConfig) error { return nil }
	workbenchLoggingSyncFn = func() error { return nil }
	workbenchStartupErrWriter = io.Discard
	workbenchAppBuilderFn = func(got *bootstrap.Result) (*app.App, error) {
		assert.Same(t, boot, got)
		return &app.App{Store: store}, nil
	}
	startAppFn = func(*app.App, context.Context) error { return nil }
	stopAppFn = func(*app.App, context.Context) error {
		stopCalls++
		return nil
	}

	err := runWorkbench("review")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create initial session: wave46 workbench create refused")
	assert.Equal(t, 1, stopCalls)
	require.Len(t, store.endKeys, 1)
	assert.Contains(t, store.endKeys[0], "workbench-")
}
