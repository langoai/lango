package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/toolcatalog"
)

type wave51SessionStore struct {
	stubCockpitSessionStore

	createdModes []string
	endKeys      []string
}

func (s *wave51SessionStore) Create(sess *session.Session) error {
	s.createdModes = append(s.createdModes, sess.Mode())
	return nil
}

func (s *wave51SessionStore) End(key string) error {
	s.endKeys = append(s.endKeys, key)
	return nil
}

type wave51ConfigProfileStore struct {
	saveCalls []wave51ConfigSaveCall
}

type wave51ConfigSaveCall struct {
	profile      string
	provider     string
	explicitKeys map[string]bool
}

func (s *wave51ConfigProfileStore) Save(_ context.Context, name string, cfg *config.Config, explicitKeys map[string]bool) error {
	s.saveCalls = append(s.saveCalls, wave51ConfigSaveCall{
		profile:      name,
		provider:     cfg.Agent.Provider,
		explicitKeys: explicitKeys,
	})
	return nil
}

func (*wave51ConfigProfileStore) Load(context.Context, string) (*config.Config, map[string]bool, error) {
	return config.DefaultConfig(), nil, nil
}

func (*wave51ConfigProfileStore) LoadActive(context.Context) (string, *config.Config, map[string]bool, error) {
	return "wave51", config.DefaultConfig(), nil, nil
}

func (*wave51ConfigProfileStore) SetActive(context.Context, string) error {
	return nil
}

func (*wave51ConfigProfileStore) List(context.Context) ([]configstore.ProfileInfo, error) {
	return nil, nil
}

func (*wave51ConfigProfileStore) Delete(context.Context, string) error {
	return nil
}

func (*wave51ConfigProfileStore) Exists(context.Context, string) (bool, error) {
	return true, nil
}

func TestWave51RunChatUsesProgramSeamAndCleansUpSession(t *testing.T) {
	restoreWave19TUIProfile(t)
	restore := replaceWave51TUIHooks(t)
	defer restore()

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	cfg.Modes = map[string]config.SessionMode{"focus": {Name: "focus"}}
	store := &wave51SessionStore{}
	stopCalls := 0
	runCalls := 0
	chatBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "wave51-chat"}, nil
	}
	chatAppBuilderFn = func(*bootstrap.Result) (*app.App, error) {
		return &app.App{
			Store:            store,
			EventBus:         eventbus.New(),
			ApprovalProvider: approval.NewCompositeProvider(),
		}, nil
	}
	stopAppFn = func(*app.App, context.Context) error {
		stopCalls++
		return nil
	}
	runTeaProgramFn = func(*tea.Program) (tea.Model, error) {
		runCalls++
		return nil, nil
	}

	err := runChat("focus")

	require.NoError(t, err)
	assert.Equal(t, 1, runCalls)
	assert.Equal(t, 1, stopCalls)
	assert.Equal(t, []string{"focus"}, store.createdModes)
	require.Len(t, store.endKeys, 1)
	assert.True(t, strings.HasPrefix(store.endKeys[0], "tui-"))
}

func TestWave51RunCockpitProgramErrorStopsAppAndReturnsTUIError(t *testing.T) {
	restoreWave19TUIProfile(t)
	restore := replaceWave51TUIHooks(t)
	defer restore()

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	cfg.Modes = map[string]config.SessionMode{"ops": {Name: "ops"}}
	store := &wave51SessionStore{}
	stopCalls := 0
	tuiErr := errors.New("wave51 cockpit program stopped")
	cockpitBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      cfg,
			ProfileName: "wave51-cockpit",
			Storage:     storage.NewFacade(&wave51ConfigProfileStore{}, nil),
		}, nil
	}
	cockpitAppBuilderFn = func(*bootstrap.Result, app.AppOption) (*app.App, error) {
		return &app.App{
			Store:            store,
			EventBus:         eventbus.New(),
			ToolCatalog:      toolcatalog.New(),
			ApprovalProvider: approval.NewCompositeProvider(),
		}, nil
	}
	stopAppFn = func(*app.App, context.Context) error {
		stopCalls++
		return nil
	}
	runTeaProgramFn = func(*tea.Program) (tea.Model, error) {
		return nil, tuiErr
	}

	err := runCockpit("ops")

	require.ErrorIs(t, err, tuiErr)
	assert.Contains(t, err.Error(), "TUI: wave51 cockpit program stopped")
	assert.Equal(t, 1, stopCalls)
	assert.Equal(t, []string{"ops"}, store.createdModes)
	require.Len(t, store.endKeys, 1)
	assert.True(t, strings.HasPrefix(store.endKeys[0], "cockpit-"))
}

func TestWave51RunWorkbenchProgramErrorStopsAppAndReturnsTUIError(t *testing.T) {
	restoreWave19TUIProfile(t)
	restore := replaceWave51TUIHooks(t)
	defer restore()

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	store := &wave51SessionStore{}
	stopCalls := 0
	tuiErr := errors.New("wave51 workbench program stopped")
	workbenchBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      cfg,
			ProfileName: "wave51-workbench",
			Storage:     storage.NewFacade(&wave51ConfigProfileStore{}, nil),
		}, nil
	}
	workbenchAppBuilderFn = func(*bootstrap.Result) (*app.App, error) {
		return &app.App{
			Store:            store,
			EventBus:         eventbus.New(),
			ToolCatalog:      toolcatalog.New(),
			ApprovalProvider: approval.NewCompositeProvider(),
		}, nil
	}
	stopAppFn = func(*app.App, context.Context) error {
		stopCalls++
		return nil
	}
	runTeaProgramFn = func(*tea.Program) (tea.Model, error) {
		return nil, tuiErr
	}

	err := runWorkbench("")

	require.ErrorIs(t, err, tuiErr)
	assert.Contains(t, err.Error(), "TUI: wave51 workbench program stopped")
	assert.Equal(t, 1, stopCalls)
	assert.Empty(t, store.createdModes)
	require.Len(t, store.endKeys, 1)
	assert.True(t, strings.HasPrefix(store.endKeys[0], "workbench-"))
}

func TestWave51ConfigGetAndSetUseCommandBootSeams(t *testing.T) {
	origBootResult := configBootResultFn
	origConfig := configConfigFn
	t.Cleanup(func() {
		configBootResultFn = origBootResult
		configConfigFn = origConfig
	})

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "wave51-provider"
	configConfigFn = func() (*config.Config, error) {
		return cfg, nil
	}
	var getOut bytes.Buffer
	getCmd := configCmd()
	getCmd.SetArgs([]string{"get", "agent.provider"})
	getCmd.SetOut(&getOut)
	getCmd.SetErr(io.Discard)

	require.NoError(t, getCmd.Execute())
	assert.Equal(t, "wave51-provider\n", getOut.String())

	profiles := &wave51ConfigProfileStore{}
	configBootResultFn = func() (*bootstrap.Result, error) {
		setCfg := config.DefaultConfig()
		return &bootstrap.Result{
			Config:      setCfg,
			ProfileName: "wave51-profile",
			Storage:     storage.NewFacade(profiles, nil),
		}, nil
	}
	var setOut bytes.Buffer
	setCmd := configCmd()
	setCmd.SetArgs([]string{"set", "agent.provider", "wave51-set"})
	setCmd.SetOut(&setOut)
	setCmd.SetErr(io.Discard)

	require.NoError(t, setCmd.Execute())
	assert.Equal(t, "Set agent.provider = wave51-set\n", setOut.String())
	require.Len(t, profiles.saveCalls, 1)
	assert.Equal(t, "wave51-profile", profiles.saveCalls[0].profile)
	assert.Equal(t, "wave51-set", profiles.saveCalls[0].provider)
	assert.Nil(t, profiles.saveCalls[0].explicitKeys)
}

func replaceWave51TUIHooks(t *testing.T) func() {
	t.Helper()
	origChatBootLoader := chatBootLoaderFn
	origChatLoggingInit := chatLoggingInitFn
	origChatLoggingSync := chatLoggingSyncFn
	origChatWriter := chatStartupErrWriter
	origChatBuilder := chatAppBuilderFn
	origCockpitBootLoader := cockpitBootLoaderFn
	origCockpitLoggingInit := cockpitLoggingInitFn
	origCockpitLoggingSync := cockpitLoggingSyncFn
	origCockpitWriter := cockpitStartupErrWriter
	origCockpitBuilder := cockpitAppBuilderFn
	origWorkbenchBootLoader := workbenchBootLoaderFn
	origWorkbenchLoggingInit := workbenchLoggingInitFn
	origWorkbenchLoggingSync := workbenchLoggingSyncFn
	origWorkbenchWriter := workbenchStartupErrWriter
	origWorkbenchBuilder := workbenchAppBuilderFn
	origStart := startAppFn
	origStop := stopAppFn
	origRunTeaProgram := runTeaProgramFn
	origWithChannels := withChannels

	chatLoggingInitFn = func(logging.LogConfig) error { return nil }
	chatLoggingSyncFn = func() error { return nil }
	chatStartupErrWriter = io.Discard
	cockpitLoggingInitFn = func(logging.LogConfig) error { return nil }
	cockpitLoggingSyncFn = func() error { return nil }
	cockpitStartupErrWriter = io.Discard
	workbenchLoggingInitFn = func(logging.LogConfig) error { return nil }
	workbenchLoggingSyncFn = func() error { return nil }
	workbenchStartupErrWriter = io.Discard
	startAppFn = func(*app.App, context.Context) error { return nil }
	withChannels = false

	return func() {
		chatBootLoaderFn = origChatBootLoader
		chatLoggingInitFn = origChatLoggingInit
		chatLoggingSyncFn = origChatLoggingSync
		chatStartupErrWriter = origChatWriter
		chatAppBuilderFn = origChatBuilder
		cockpitBootLoaderFn = origCockpitBootLoader
		cockpitLoggingInitFn = origCockpitLoggingInit
		cockpitLoggingSyncFn = origCockpitLoggingSync
		cockpitStartupErrWriter = origCockpitWriter
		cockpitAppBuilderFn = origCockpitBuilder
		workbenchBootLoaderFn = origWorkbenchBootLoader
		workbenchLoggingInitFn = origWorkbenchLoggingInit
		workbenchLoggingSyncFn = origWorkbenchLoggingSync
		workbenchStartupErrWriter = origWorkbenchWriter
		workbenchAppBuilderFn = origWorkbenchBuilder
		startAppFn = origStart
		stopAppFn = origStop
		runTeaProgramFn = origRunTeaProgram
		withChannels = origWithChannels
	}
}
