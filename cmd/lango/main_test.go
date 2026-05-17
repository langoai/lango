package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/cliexit"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/cli/cockpit/pages"
	cliextension "github.com/langoai/lango/internal/cli/extension"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/collabview"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/mcp"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/proposal"
	"github.com/langoai/lango/internal/session"
)

type fakeServeApp struct {
	stopFn func(ctx context.Context) error
}

func (f *fakeServeApp) Start(ctx context.Context) error { return nil }

func (f *fakeServeApp) Stop(ctx context.Context) error {
	if f.stopFn != nil {
		return f.stopFn(ctx)
	}
	return nil
}

type stubCockpitSessionStore struct{}

func (stubCockpitSessionStore) Create(*session.Session) error               { return nil }
func (stubCockpitSessionStore) Get(string) (*session.Session, error)        { return nil, nil }
func (stubCockpitSessionStore) Update(*session.Session) error               { return nil }
func (stubCockpitSessionStore) Delete(string) error                         { return nil }
func (stubCockpitSessionStore) AppendMessage(string, session.Message) error { return nil }
func (stubCockpitSessionStore) AnnotateTimeout(string, string) error        { return nil }
func (stubCockpitSessionStore) End(string) error                            { return nil }
func (stubCockpitSessionStore) Close() error                                { return nil }
func (stubCockpitSessionStore) ListSessions(context.Context) ([]session.SessionSummary, error) {
	return nil, nil
}
func (stubCockpitSessionStore) GetSalt(string) ([]byte, error) { return nil, nil }
func (stubCockpitSessionStore) SetSalt(string, []byte) error   { return nil }

func TestWatchServeSignals_FirstSignalStartsGracefulShutdown(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopped := make(chan struct{})
	app := &fakeServeApp{
		stopFn: func(ctx context.Context) error {
			close(stopped)
			return nil
		},
	}

	sigChan := make(chan os.Signal, 2)
	forced := make(chan int, 1)

	go watchServeSignals(ctx, app, zap.NewNop().Sugar(), sigChan, time.Second, cancel, func(code int) {
		forced <- code
	})

	sigChan <- os.Interrupt

	select {
	case <-stopped:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected graceful shutdown to start")
	}

	select {
	case code := <-forced:
		t.Fatalf("unexpected forced exit with code %d", code)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestWatchServeSignals_SecondSignalForcesExit(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	release := make(chan struct{})
	app := &fakeServeApp{
		stopFn: func(ctx context.Context) error {
			<-release
			return nil
		},
	}

	sigChan := make(chan os.Signal, 2)
	forced := make(chan int, 1)
	var once sync.Once

	go watchServeSignals(ctx, app, zap.NewNop().Sugar(), sigChan, time.Second, cancel, func(code int) {
		once.Do(func() { forced <- code })
	})

	sigChan <- os.Interrupt
	sigChan <- os.Interrupt

	select {
	case code := <-forced:
		assert.Equal(t, 130, code)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected forced exit on second signal")
	}

	close(release)
}

func TestCockpitDeadLetterListOptions_MapsAllFields(t *testing.T) {
	t.Parallel()

	got := cockpitDeadLetterListOptions(pages.DeadLetterListOptions{
		Query:                     "tx-1",
		Adjudication:              "release",
		LatestStatusSubtype:       "dead-lettered",
		LatestStatusSubtypeFamily: "dead-letter",
		AnyMatchFamily:            "manual-retry",
		ManualReplayActor:         "operator:alice",
		DeadLetteredAfter:         "2026-04-27T10:00:00Z",
		DeadLetteredBefore:        "2026-04-27T11:00:00Z",
		DeadLetterReasonQuery:     "worker exhausted",
		LatestDispatchReference:   "dispatch-7",
	})

	assert.Equal(t, cockpit.DeadLetterListOptions{
		Query:                     "tx-1",
		Adjudication:              "release",
		LatestStatusSubtype:       "dead-lettered",
		LatestStatusSubtypeFamily: "dead-letter",
		AnyMatchFamily:            "manual-retry",
		ManualReplayActor:         "operator:alice",
		DeadLetteredAfter:         "2026-04-27T10:00:00Z",
		DeadLetteredBefore:        "2026-04-27T11:00:00Z",
		DeadLetterReasonQuery:     "worker exhausted",
		LatestDispatchReference:   "dispatch-7",
	}, got)
}

func TestRunCockpitBuildDepsCarriesMissionService(t *testing.T) {
	t.Parallel()

	svc := mission.NewService(nil)
	store := &stubMainMissionStore{}
	registry := proposal.NewRegistry(nil)
	psvc := proposal.NewService(registry, nil)
	inquiryReader := &stubMainLoopInquiryReader{}
	deadReader := &stubMainLoopDeadReader{}
	cronReader := &stubMainLoopCronReader{}
	collabMissionLinks := &stubMainCollabMissionLinks{}
	collabAgentRuns := &stubMainCollabAgentRuns{}
	collabDelegations := &stubMainCollabDelegations{}
	collabRuntime := &stubMainCollabRuntime{}
	application := &app.App{
		MissionService:                 svc,
		MissionStore:                   store,
		ProposalRegistry:               registry,
		ProposalService:                psvc,
		LoopInquiryReader:              inquiryReader,
		LoopDeadLetterReader:           deadReader,
		LoopCronReader:                 cronReader,
		CollaborationMissionLinkReader: collabMissionLinks,
		CollaborationAgentRunReader:    collabAgentRuns,
		CollaborationDelegationReader:  collabDelegations,
		CollaborationRuntimeReader:     collabRuntime,
	}
	cfg := &config.Config{}
	pending := cockpit.NewPendingApprovalRegistry()
	learning := cockpit.NewLearningSuggestionBuffer(nil)
	activity := cockpit.NewMissionActivityBuffer()

	deps := buildMissionControlDeps(application, cfg, "", "sess-1", nil, "", nil, pending, learning, activity)

	assert.Same(t, svc, deps.MissionService)
	assert.Same(t, store, deps.MissionReader)
	assert.Same(t, registry, deps.ProposalReader)
	assert.Same(t, psvc, deps.ProposalService)
	assert.Same(t, inquiryReader, deps.LoopInquiryReader)
	assert.Same(t, deadReader, deps.LoopDeadReader)
	assert.Same(t, cronReader, deps.LoopCronReader)
	assert.Same(t, collabMissionLinks, deps.CollabMissionLinks)
	assert.Same(t, collabAgentRuns, deps.CollabAgentRuns)
	assert.Same(t, collabDelegations, deps.CollabDelegations)
	assert.Same(t, collabRuntime, deps.CollabRuntime)
	assert.Same(t, learning, deps.LearningBuffer)
	assert.Same(t, activity, deps.ActivityBuffer)
}

func TestBuildMissionControlDepsCarriesChatRuntimeFeatures(t *testing.T) {
	t.Parallel()

	application := &app.App{
		MCPManager: mcp.NewServerManager(config.MCPConfig{}),
	}
	deps := buildMissionControlDeps(
		application,
		&config.Config{},
		"",
		"sess-1",
		nil,
		"",
		nil,
		cockpit.NewPendingApprovalRegistry(),
		cockpit.NewLearningSuggestionBuffer(nil),
		cockpit.NewMissionActivityBuffer(),
	)

	assert.True(t, deps.RuntimeFeatures.MCPActive)
}

func TestRegisterCockpitPages_AlwaysRegistersStatusAndSettings(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	application := &app.App{
		Store: stubCockpitSessionStore{},
	}
	model := cockpit.New(cockpit.Deps{})
	chatModel := model.ChatModel()
	require.NotNil(t, chatModel)

	registerCockpitPages(
		model,
		application,
		cfg,
		"",
		nil,
		cockpit.Deps{},
		chatModel,
	)

	_, hasStatus := model.Pages()[cockpit.PageStatus]
	_, hasSettings := model.Pages()[cockpit.PageSettings]
	require.True(t, hasStatus, "status page should always be registered")
	require.True(t, hasSettings, "settings page should always be registered")
	assert.False(t, model.Sidebar().IsDisabled(cockpit.PageStatus.String()))
	assert.False(t, model.Sidebar().IsDisabled(cockpit.PageSettings.String()))
}

func TestRegisterCockpitPages_AlwaysRegistersDeadLetters(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	application := &app.App{
		Store: stubCockpitSessionStore{},
	}
	model := cockpit.New(cockpit.Deps{})
	chatModel := model.ChatModel()
	require.NotNil(t, chatModel)

	registerCockpitPages(
		model,
		application,
		cfg,
		"",
		nil,
		cockpit.Deps{},
		chatModel,
	)

	_, hasDeadLetters := model.Pages()[cockpit.PageDeadLetters]
	require.True(t, hasDeadLetters, "dead letters page should always be registered")
	assert.False(t, model.Sidebar().IsDisabled(cockpit.PageDeadLetters.String()))
}

func TestNewRootCmdRoutesInteractiveRootToWorkbench(t *testing.T) {
	prevInteractive := isInteractiveFn
	prevWorkbench := runWorkbenchFn
	defer func() {
		isInteractiveFn = prevInteractive
		runWorkbenchFn = prevWorkbench
	}()

	isInteractiveFn = func() bool { return true }
	called := false
	gotMode := ""
	runWorkbenchFn = func(mode string) error {
		called = true
		gotMode = mode
		return nil
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--mode", "research"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "research", gotMode)
}

func TestCockpitCmdRoutesToExplicitCockpitRunner(t *testing.T) {
	prevInteractive := isInteractiveFn
	prevCockpit := runCockpitFn
	defer func() {
		isInteractiveFn = prevInteractive
		runCockpitFn = prevCockpit
	}()

	isInteractiveFn = func() bool { return true }
	called := false
	gotMode := ""
	runCockpitFn = func(mode string) error {
		called = true
		gotMode = mode
		return nil
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"cockpit", "--mode", "debug"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "debug", gotMode)
}

func TestChatCmdRoutesToExplicitChatRunner(t *testing.T) {
	prevInteractive := isInteractiveFn
	prevChat := runChatFn
	defer func() {
		isInteractiveFn = prevInteractive
		runChatFn = prevChat
	}()

	isInteractiveFn = func() bool { return true }
	called := false
	gotMode := ""
	runChatFn = func(mode string) error {
		called = true
		gotMode = mode
		return nil
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"chat", "--mode", "review"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "review", gotMode)
}

func TestCockpitCmd_NonInteractiveReturnsActionableError(t *testing.T) {
	prevInteractive := isInteractiveFn
	defer func() { isInteractiveFn = prevInteractive }()

	isInteractiveFn = func() bool { return false }

	cmd := newRootCmd()
	cmd.SetArgs([]string{"cockpit"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cockpit requires an interactive terminal")
}

func TestChatCmd_NonInteractiveReturnsActionableError(t *testing.T) {
	prevInteractive := isInteractiveFn
	defer func() { isInteractiveFn = prevInteractive }()

	isInteractiveFn = func() bool { return false }

	cmd := newRootCmd()
	cmd.SetArgs([]string{"chat"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chat requires an interactive terminal")
}

func TestNewRootCmd_NonInteractiveBareRootWritesHelpToCommandOutput(t *testing.T) {
	prevInteractive := isInteractiveFn
	defer func() { isInteractiveFn = prevInteractive }()

	isInteractiveFn = func() bool { return false }

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(nil)

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Lango is a high-performance AI agent built with Go")
	assert.Contains(t, out.String(), "Usage:")
}

func TestNewRootCmdBgListReportsStandaloneBoundary(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"bg", "list"})

	err := cmd.Execute()
	require.Error(t, err)

	msg := err.Error()
	assert.Contains(t, msg, "background task state is in-memory")
	assert.Contains(t, msg, "owned by the running app/server process")
	assert.Contains(t, msg, "standalone root CLI is not yet connected")
	assert.Contains(t, msg, "gateway API")
	assert.NotContains(t, msg, "lango serve")
}

func TestNewRootCmd_InvalidModeReturnsActionableError(t *testing.T) {
	prevInteractive := isInteractiveFn
	prevRunWorkbench := runWorkbenchFn
	defer func() {
		isInteractiveFn = prevInteractive
		runWorkbenchFn = prevRunWorkbench
	}()

	isInteractiveFn = func() bool { return true }
	runWorkbenchFn = func(mode string) error {
		return fmt.Errorf("unknown mode %q; valid modes can be listed via /mode", mode)
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--mode", "does-not-exist"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown mode "does-not-exist"`)
}

func TestCockpitCmd_InvalidModeReturnsActionableError(t *testing.T) {
	prevInteractive := isInteractiveFn
	prevRunCockpit := runCockpitFn
	defer func() {
		isInteractiveFn = prevInteractive
		runCockpitFn = prevRunCockpit
	}()

	isInteractiveFn = func() bool { return true }
	runCockpitFn = func(mode string) error {
		return fmt.Errorf("unknown mode %q; valid modes can be listed via /mode", mode)
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"cockpit", "--mode", "does-not-exist"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown mode "does-not-exist"`)
}

func TestChatCmd_InvalidModeReturnsActionableError(t *testing.T) {
	prevInteractive := isInteractiveFn
	prevRunChat := runChatFn
	defer func() {
		isInteractiveFn = prevInteractive
		runChatFn = prevRunChat
	}()

	isInteractiveFn = func() bool { return true }
	runChatFn = func(mode string) error {
		return fmt.Errorf("unknown mode %q; valid modes can be listed via /mode", mode)
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"chat", "--mode", "does-not-exist"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown mode "does-not-exist"`)
}

func TestRunMain_BrokerModeWritesErrorToInjectedStderr(t *testing.T) {
	origSandbox := isSandboxWorkerModeFn
	origBroker := isStorageBrokerModeFn
	origRunBroker := runStorageBrokerServerFn
	origStdin := mainStdin
	origStdout := mainStdout
	origStderr := mainStderr
	t.Cleanup(func() {
		isSandboxWorkerModeFn = origSandbox
		isStorageBrokerModeFn = origBroker
		runStorageBrokerServerFn = origRunBroker
		mainStdin = origStdin
		mainStdout = origStdout
		mainStderr = origStderr
	})

	isSandboxWorkerModeFn = func() bool { return false }
	isStorageBrokerModeFn = func() bool { return true }
	runStorageBrokerServerFn = func(in io.Reader, out io.Writer) error {
		return errors.New("broker failed")
	}
	mainStdin = bytes.NewBuffer(nil)
	mainStdout = &bytes.Buffer{}
	var errBuf bytes.Buffer
	mainStderr = &errBuf

	code := runMain()

	assert.Equal(t, 1, code)
	assert.Contains(t, errBuf.String(), "broker failed")
}

func TestRunMain_BrokerModeUsesInjectedSTDIO(t *testing.T) {
	origSandbox := isSandboxWorkerModeFn
	origBroker := isStorageBrokerModeFn
	origRunBroker := runStorageBrokerServerFn
	origStdin := mainStdin
	origStdout := mainStdout
	origStderr := mainStderr
	t.Cleanup(func() {
		isSandboxWorkerModeFn = origSandbox
		isStorageBrokerModeFn = origBroker
		runStorageBrokerServerFn = origRunBroker
		mainStdin = origStdin
		mainStdout = origStdout
		mainStderr = origStderr
	})

	isSandboxWorkerModeFn = func() bool { return false }
	isStorageBrokerModeFn = func() bool { return true }

	in := bytes.NewBufferString("broker-input")
	var out bytes.Buffer
	mainStdin = in
	mainStdout = &out
	mainStderr = &bytes.Buffer{}

	called := false
	runStorageBrokerServerFn = func(gotIn io.Reader, gotOut io.Writer) error {
		called = true
		assert.Same(t, in, gotIn)
		assert.Same(t, &out, gotOut)
		_, err := gotOut.Write([]byte("broker-output"))
		return err
	}

	code := runMain()

	assert.Equal(t, 0, code)
	assert.True(t, called)
	assert.Equal(t, "broker-output", out.String())
}

func TestRunMain_WorkerModeShortCircuitsToWorkerSeam(t *testing.T) {
	origSandbox := isSandboxWorkerModeFn
	origRunWorker := runSandboxWorkerFn
	origBroker := isStorageBrokerModeFn
	origNewRoot := newRootCmdFn
	t.Cleanup(func() {
		isSandboxWorkerModeFn = origSandbox
		runSandboxWorkerFn = origRunWorker
		isStorageBrokerModeFn = origBroker
		newRootCmdFn = origNewRoot
	})

	isSandboxWorkerModeFn = func() bool { return true }
	workerCalled := false
	runSandboxWorkerFn = func() int {
		workerCalled = true
		return 7
	}

	isStorageBrokerModeFn = func() bool {
		t.Fatal("broker mode should not be checked after worker short-circuit")
		return false
	}
	newRootCmdFn = func() *cobra.Command {
		t.Fatal("root command should not be constructed after worker short-circuit")
		return nil
	}

	code := runMain()

	assert.Equal(t, 7, code)
	assert.True(t, workerCalled)
}

func TestRunMain_RootCommandErrorWritesToInjectedStderr(t *testing.T) {
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
			Use: "lango",
			RunE: func(cmd *cobra.Command, args []string) error {
				return errors.New("root execute failed")
			},
		}
	}
	var errBuf bytes.Buffer
	mainStderr = &errBuf

	code := runMain()

	assert.Equal(t, 1, code)
	assert.Contains(t, errBuf.String(), "root execute failed")
}

func TestRunMain_RootCommandStructuredExitCodeWritesToInjectedStderr(t *testing.T) {
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
				return cliexit.New(3, errors.New("stdin is not a TTY; pass --yes for scripted runs"))
			},
		}
	}
	var errBuf bytes.Buffer
	mainStderr = &errBuf

	code := runMain()

	assert.Equal(t, 3, code)
	assert.Equal(t, "stdin is not a TTY; pass --yes for scripted runs\n", errBuf.String())
}

func TestRunMain_RootCommandSilentStructuredExitCodeDoesNotWriteStderr(t *testing.T) {
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
				return cliexit.NewSilent(3)
			},
		}
	}
	var errBuf bytes.Buffer
	mainStderr = &errBuf

	code := runMain()

	assert.Equal(t, 3, code)
	assert.Empty(t, errBuf.String())
}

func TestRunMain_ExtensionStructuredExitCodeFromRootCommand(t *testing.T) {
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

	enabled := true
	cfg := &config.Config{
		Extensions: config.ExtensionsConfig{
			Enabled: &enabled,
			Dir:     filepath.Join(t.TempDir(), "extensions"),
		},
		Skill: config.SkillConfig{SkillsDir: filepath.Join(t.TempDir(), "skills")},
	}
	packDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(packDir, "extension.yaml"), []byte("schema: lango.extension/v1\n"), 0o644))

	var cobraErr bytes.Buffer
	isSandboxWorkerModeFn = func() bool { return false }
	isStorageBrokerModeFn = func() bool { return false }
	newRootCmdFn = func() *cobra.Command {
		root := &cobra.Command{Use: "lango"}
		root.SetErr(&cobraErr)
		extensionCmd := cliextension.NewExtensionCmd(func() (*config.Config, error) {
			return cfg, nil
		})
		root.AddCommand(extensionCmd)
		root.SetArgs([]string{"extension", "inspect", packDir})
		return root
	}
	var errBuf bytes.Buffer
	mainStderr = &errBuf

	code := runMain()

	assert.Equal(t, 1, code)
	assert.Empty(t, cobraErr.String(), "cobra should not duplicate structured CLI errors")
	assert.Contains(t, errBuf.String(), "invalid pack name")
	assert.Equal(t, 1, strings.Count(errBuf.String(), "invalid pack name"))
}

func TestVersionCmd_WritesToCommandOutput(t *testing.T) {
	origVersion := Version
	origBuildTime := BuildTime
	t.Cleanup(func() {
		Version = origVersion
		BuildTime = origBuildTime
	})

	Version = "9.9.9-test"
	BuildTime = "2026-05-14T00:00:00Z"

	cmd := versionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "lango 9.9.9-test (built 2026-05-14T00:00:00Z)\n", out.String())
}

func TestVersionCmd_IgnoresRootModeFlag(t *testing.T) {
	origVersion := Version
	origBuildTime := BuildTime
	t.Cleanup(func() {
		Version = origVersion
		BuildTime = origBuildTime
	})

	Version = "9.9.9-test"
	BuildTime = "2026-05-14T00:00:00Z"

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"version", "--mode", "research"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "lango 9.9.9-test (built 2026-05-14T00:00:00Z)\n", out.String())
}

func TestHealthCmd_WritesToCommandOutput(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	})

	port := listener.Addr().(*net.TCPAddr).Port

	cmd := healthCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--port", strconv.Itoa(port)})

	err = cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "ok\n", out.String())
}

func TestHealthCmd_IgnoresRootModeFlag(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	})

	port := listener.Addr().(*net.TCPAddr).Port

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"health", "--mode", "research", "--port", strconv.Itoa(port)})

	err = cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "ok\n", out.String())
}

func TestHealthCmd_Non200ReturnsActionableErrorWithoutSuccessOutput(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	})

	port := listener.Addr().(*net.TCPAddr).Port

	cmd := healthCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--port", strconv.Itoa(port)})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unhealthy: status 503")
	assert.Empty(t, out.String())
}

func TestHealthCmd_TimeoutReturnsErrorWithoutSuccessOutput(t *testing.T) {
	origClientFn := newHealthHTTPClientFn
	t.Cleanup(func() { newHealthHTTPClientFn = origClientFn })
	newHealthHTTPClientFn = func() *http.Client {
		return &http.Client{Timeout: 10 * time.Millisecond}
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	})

	port := listener.Addr().(*net.TCPAddr).Port

	cmd := healthCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--port", strconv.Itoa(port)})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "health check:")
	assert.Empty(t, out.String())
}

func TestRunWorkbench_InvalidModeRejectsBeforeAppBuild(t *testing.T) {
	origBootLoader := workbenchBootLoaderFn
	origAppBuilder := workbenchAppBuilderFn
	t.Cleanup(func() {
		workbenchBootLoaderFn = origBootLoader
		workbenchAppBuilderFn = origAppBuilder
	})

	workbenchBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      config.DefaultConfig(),
			ProfileName: "test",
		}, nil
	}
	workbenchAppBuilderFn = func(*bootstrap.Result) (*app.App, error) {
		t.Fatal("app builder must not run when mode validation fails")
		return nil, nil
	}

	err := runWorkbench("does-not-exist")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown mode "does-not-exist"`)
}

func TestRunCockpit_InvalidModeRejectsBeforeAppBuild(t *testing.T) {
	origBootLoader := cockpitBootLoaderFn
	origAppBuilder := cockpitAppBuilderFn
	t.Cleanup(func() {
		cockpitBootLoaderFn = origBootLoader
		cockpitAppBuilderFn = origAppBuilder
	})

	cockpitBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      config.DefaultConfig(),
			ProfileName: "test",
		}, nil
	}
	cockpitAppBuilderFn = func(*bootstrap.Result, app.AppOption) (*app.App, error) {
		t.Fatal("app builder must not run when mode validation fails")
		return nil, nil
	}

	err := runCockpit("does-not-exist")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown mode "does-not-exist"`)
}

func TestRunChat_InvalidModeRejectsBeforeAppBuild(t *testing.T) {
	origBootLoader := chatBootLoaderFn
	origAppBuilder := chatAppBuilderFn
	t.Cleanup(func() {
		chatBootLoaderFn = origBootLoader
		chatAppBuilderFn = origAppBuilder
	})

	chatBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      config.DefaultConfig(),
			ProfileName: "test",
		}, nil
	}
	chatAppBuilderFn = func(*bootstrap.Result) (*app.App, error) {
		t.Fatal("app builder must not run when mode validation fails")
		return nil, nil
	}

	err := runChat("does-not-exist")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown mode "does-not-exist"`)
}

func TestServeCmd_WritesBannerAndSummaryToCommandOutput(t *testing.T) {
	origBootLoader := serveBootLoaderFn
	origLoggingInit := serveLoggingInitFn
	origLoggingSync := serveLoggingSyncFn
	origAppBuilder := serveAppBuilderFn
	origAwaitShutdown := serveAwaitShutdownFn
	t.Cleanup(func() {
		serveBootLoaderFn = origBootLoader
		serveLoggingInitFn = origLoggingInit
		serveLoggingSyncFn = origLoggingSync
		serveAppBuilderFn = origAppBuilder
		serveAwaitShutdownFn = origAwaitShutdown
	})

	cfg := config.DefaultConfig()
	serveBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "default"}, nil
	}
	serveLoggingInitFn = func(logging.LogConfig) error { return nil }
	serveLoggingSyncFn = func() error { return nil }
	serveAppBuilderFn = func(boot *bootstrap.Result) (stoppableApplication, error) {
		return &fakeServeApp{}, nil
	}
	serveAwaitShutdownFn = func(ctx context.Context) {}

	cmd := serveCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), tui.ServeBanner())
	assert.Contains(t, out.String(), startupSummary(cfg))
	assert.Empty(t, errOut.String())
}

func TestRunCockpit_WritesStartupNoticeToInjectedStderr(t *testing.T) {
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
	cockpitBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "default"}, nil
	}
	cockpitLoggingInitFn = func(logging.LogConfig) error { return nil }
	cockpitLoggingSyncFn = func() error { return nil }
	var errBuf bytes.Buffer
	cockpitStartupErrWriter = &errBuf
	cockpitAppBuilderFn = func(boot *bootstrap.Result, mode app.AppOption) (*app.App, error) {
		return nil, errors.New("app build failed")
	}

	err := runCockpit("")
	require.Error(t, err)
	assert.Contains(t, errBuf.String(), tui.Banner())
	assert.Contains(t, errBuf.String(), "Initializing cockpit...")
}

func TestRunWorkbench_WritesStartupNoticeToInjectedStderr(t *testing.T) {
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
	workbenchBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "default"}, nil
	}
	workbenchLoggingInitFn = func(logging.LogConfig) error { return nil }
	workbenchLoggingSyncFn = func() error { return nil }
	var errBuf bytes.Buffer
	workbenchStartupErrWriter = &errBuf
	workbenchAppBuilderFn = func(boot *bootstrap.Result) (*app.App, error) {
		return nil, errors.New("app build failed")
	}

	err := runWorkbench("")
	require.Error(t, err)
	assert.Contains(t, errBuf.String(), tui.Banner())
	assert.Contains(t, errBuf.String(), "Initializing workbench...")
}

func TestRunChat_WritesStartupNoticeToInjectedStderr(t *testing.T) {
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
	chatBootLoaderFn = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "default"}, nil
	}
	chatLoggingInitFn = func(logging.LogConfig) error { return nil }
	chatLoggingSyncFn = func() error { return nil }
	var errBuf bytes.Buffer
	chatStartupErrWriter = &errBuf
	chatAppBuilderFn = func(boot *bootstrap.Result) (*app.App, error) {
		return nil, errors.New("app build failed")
	}

	err := runChat("")
	require.Error(t, err)
	assert.Contains(t, errBuf.String(), tui.Banner())
	assert.Contains(t, errBuf.String(), "Initializing...")
}

func TestCockpitCommandHelpTextNoLongerClaimsBareEquivalence(t *testing.T) {
	cmd := cockpitCmd()
	assert.NotContains(t, cmd.Short, "same as bare lango")
	assert.Contains(t, cmd.Short, "operator dashboard")
}

type stubMainMissionStore struct{}

func (*stubMainMissionStore) CreateMission(context.Context, mission.CreateMissionInput) (*mission.Mission, error) {
	return nil, nil
}
func (*stubMainMissionStore) GetMission(context.Context, string) (*mission.Mission, error) {
	return nil, nil
}
func (*stubMainMissionStore) ListMissionsBySession(context.Context, string, int) ([]*mission.Mission, error) {
	return nil, nil
}
func (*stubMainMissionStore) TransitionMission(context.Context, mission.TransitionMissionInput) (*mission.Mission, error) {
	return nil, nil
}
func (*stubMainMissionStore) AppendExecutionLink(context.Context, mission.AppendExecutionLinkInput) error {
	return nil
}
func (*stubMainMissionStore) ListExecutionLinks(context.Context, string) ([]*mission.ExecutionLink, error) {
	return nil, nil
}
func (*stubMainMissionStore) FindExecutionLinkByExecution(context.Context, mission.ExecutionKind, string) (*mission.ExecutionLink, error) {
	return nil, nil
}
func (*stubMainMissionStore) FindMissionByExecution(context.Context, mission.ExecutionKind, string) (*mission.Mission, error) {
	return nil, nil
}

type stubMainLoopInquiryReader struct{}

func (*stubMainLoopInquiryReader) ListPendingInquiries(context.Context, string, int) ([]librarian.Inquiry, error) {
	return nil, nil
}

type stubMainLoopDeadReader struct{}

func (*stubMainLoopDeadReader) ListCurrentDeadLetters(context.Context) ([]postadjudicationstatus.DeadLetterBacklogEntry, error) {
	return nil, nil
}

type stubMainLoopCronReader struct{}

func (*stubMainLoopCronReader) List(context.Context) ([]cron.Job, error) {
	return []cron.Job{{ID: uuid.NewString(), Name: "job"}}, nil
}

func (*stubMainLoopCronReader) ListHistory(context.Context, string, int) ([]cron.HistoryEntry, error) {
	return nil, nil
}

type stubMainCollabMissionLinks struct{}

func (*stubMainCollabMissionLinks) ListMissionExecutionLinks(context.Context, string) ([]collabview.CollaborationMissionExecutionLink, error) {
	return nil, nil
}

type stubMainCollabAgentRuns struct{}

func (*stubMainCollabAgentRuns) ListAgentRuns() []collabview.CollaborationAgentRunView { return nil }

type stubMainCollabDelegations struct{}

func (*stubMainCollabDelegations) ListDelegationsForSession(context.Context, string) ([]collabview.CollaborationDelegationRecord, error) {
	return nil, nil
}

type stubMainCollabRuntime struct{}

func (*stubMainCollabRuntime) ListBudgetSignals(string) []collabview.CollaborationBudgetRecord {
	return nil
}
func (*stubMainCollabRuntime) ListRecoverySignals(string) []collabview.CollaborationRecoveryRecord {
	return nil
}
