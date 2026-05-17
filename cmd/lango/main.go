package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
	clia2a "github.com/langoai/lango/internal/cli/a2a"
	cliagent "github.com/langoai/lango/internal/cli/agent"
	clialerts "github.com/langoai/lango/internal/cli/alerts"
	cliapproval "github.com/langoai/lango/internal/cli/approval"
	clibg "github.com/langoai/lango/internal/cli/bg"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/cli/cliboot"
	"github.com/langoai/lango/internal/cli/cliexit"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/cli/cockpit/pages"
	cliconfigcmd "github.com/langoai/lango/internal/cli/configcmd"
	clicontract "github.com/langoai/lango/internal/cli/contract"
	clicron "github.com/langoai/lango/internal/cli/cron"
	"github.com/langoai/lango/internal/cli/doctor"
	clieconomy "github.com/langoai/lango/internal/cli/economy"
	cliextension "github.com/langoai/lango/internal/cli/extension"
	cligraph "github.com/langoai/lango/internal/cli/graph"
	clilearning "github.com/langoai/lango/internal/cli/learning"
	clilibrarian "github.com/langoai/lango/internal/cli/librarian"
	climcp "github.com/langoai/lango/internal/cli/mcp"
	climemory "github.com/langoai/lango/internal/cli/memory"
	climetrics "github.com/langoai/lango/internal/cli/metrics"
	"github.com/langoai/lango/internal/cli/onboard"
	clip2p "github.com/langoai/lango/internal/cli/p2p"
	clipayment "github.com/langoai/lango/internal/cli/payment"
	"github.com/langoai/lango/internal/cli/prompt"
	cliprovenance "github.com/langoai/lango/internal/cli/provenance"
	clirun "github.com/langoai/lango/internal/cli/run"
	clisandbox "github.com/langoai/lango/internal/cli/sandbox"
	clisecurity "github.com/langoai/lango/internal/cli/security"
	"github.com/langoai/lango/internal/cli/settings"
	cliaccount "github.com/langoai/lango/internal/cli/smartaccount"
	clistatus "github.com/langoai/lango/internal/cli/status"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/cli/workbench"
	cliworkflow "github.com/langoai/lango/internal/cli/workflow"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/sandbox"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/storagebroker"
	"github.com/langoai/lango/internal/types"
	"go.uber.org/zap"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

var (
	exitFn                             = os.Exit
	runWorkbenchFn                     = runWorkbench
	runCockpitFn                       = runCockpit
	runChatFn                          = runChat
	isInteractiveFn                    = prompt.IsInteractive
	mainStdin                io.Reader = os.Stdin
	mainStdout               io.Writer = os.Stdout
	mainStderr               io.Writer = os.Stderr
	isSandboxWorkerModeFn              = sandbox.IsWorkerMode
	runSandboxWorkerFn                 = func() { sandbox.RunWorker(sandbox.ToolRegistry{}) }
	isStorageBrokerModeFn              = storagebroker.IsBrokerMode
	runStorageBrokerServerFn           = func(in io.Reader, out io.Writer) error {
		return storagebroker.NewServer().Run(in, out)
	}
	newRootCmdFn       = newRootCmd
	serveBootLoaderFn  = cliboot.BootResult
	serveLoggingInitFn = logging.Init
	serveLoggingSyncFn = logging.Sync
	serveAppBuilderFn  = func(boot *bootstrap.Result) (stoppableApplication, error) {
		return app.New(boot)
	}
	serveAwaitShutdownFn = func(ctx context.Context) {
		<-ctx.Done()
	}
	newHealthHTTPClientFn = func() *http.Client {
		return &http.Client{Timeout: 5 * time.Second}
	}
	cockpitBootLoaderFn                 = cliboot.BootResult
	cockpitLoggingInitFn                = logging.Init
	cockpitLoggingSyncFn                = logging.Sync
	cockpitStartupErrWriter   io.Writer = os.Stderr
	cockpitAppBuilderFn                 = func(boot *bootstrap.Result, mode app.AppOption) (*app.App, error) { return app.New(boot, mode) }
	workbenchBootLoaderFn               = cliboot.BootResult
	workbenchLoggingInitFn              = logging.Init
	workbenchLoggingSyncFn              = logging.Sync
	workbenchStartupErrWriter io.Writer = os.Stderr
	workbenchAppBuilderFn               = func(boot *bootstrap.Result) (*app.App, error) { return app.New(boot, app.WithLocalChat()) }
	chatBootLoaderFn                    = cliboot.BootResult
	chatLoggingInitFn                   = logging.Init
	chatLoggingSyncFn                   = logging.Sync
	chatStartupErrWriter      io.Writer = os.Stderr
	chatAppBuilderFn                    = func(boot *bootstrap.Result) (*app.App, error) { return app.New(boot, app.WithLocalChat()) }
)

type stoppableApplication interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func main() {
	if code := runMain(); code != 0 {
		exitFn(code)
	}
}

func runMain() int {
	// Check if running as sandbox worker subprocess.
	if isSandboxWorkerModeFn() {
		runSandboxWorkerFn()
		return 0
	}
	if isStorageBrokerModeFn() {
		if err := runStorageBrokerServerFn(mainStdin, mainStdout); err != nil {
			fmt.Fprintln(mainStderr, err)
			return 1
		}
		return 0
	}

	tui.SetVersionInfo(Version, BuildTime)
	cliboot.Version = Version

	rootCmd := newRootCmdFn()
	if err := rootCmd.Execute(); err != nil {
		if code, ok := cliexit.Code(err); ok {
			if !cliexit.Silent(err) {
				fmt.Fprintln(mainStderr, err)
			}
			return code
		}
		fmt.Fprintln(mainStderr, err)
		return 1
	}
	return 0
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "lango",
		Short: "Lango - Fast AI Agent in Go",
		Long:  `Lango is a high-performance AI agent built with Go, supporting multiple channels and tools.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isInteractiveFn() {
				return cmd.Help()
			}
			modeName, _ := cmd.Flags().GetString("mode")
			return runWorkbenchFn(modeName)
		},
	}
	rootCmd.PersistentFlags().String("mode", "", "Initial session mode (e.g., code-review, research, debug)")

	rootCmd.AddGroup(
		&cobra.Group{ID: "start", Title: "Getting Started:"},
		&cobra.Group{ID: "ai", Title: "AI & Knowledge:"},
		&cobra.Group{ID: "auto", Title: "Automation:"},
		&cobra.Group{ID: "net", Title: "Network & Economy:"},
		&cobra.Group{ID: "sys", Title: "Security & System:"},
	)

	// --- Getting Started ---
	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(cockpitCmd())
	rootCmd.AddCommand(chatCmd())

	onboardCmd := onboard.NewCommand()
	onboardCmd.GroupID = "start"
	rootCmd.AddCommand(onboardCmd)

	doctorCmd := doctor.NewCommand()
	doctorCmd.GroupID = "start"
	rootCmd.AddCommand(doctorCmd)

	settingsCmd := settings.NewCommand()
	settingsCmd.GroupID = "start"
	rootCmd.AddCommand(settingsCmd)

	statusCmd := clistatus.NewStatusCmd(cliboot.BootResult, newDeadLetterStatusLoader(cliboot.BootResult))
	statusCmd.GroupID = "start"
	rootCmd.AddCommand(statusCmd)

	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(configCmd())

	// --- Security & System ---
	securityCmd := clisecurity.NewSecurityCmd(cliboot.BootResult)
	securityCmd.GroupID = "sys"
	rootCmd.AddCommand(securityCmd)

	// --- AI & Knowledge ---
	memoryCmd := climemory.NewMemoryCmd(cliboot.Config)
	memoryCmd.GroupID = "ai"
	rootCmd.AddCommand(memoryCmd)

	agentCmd := cliagent.NewAgentCmd(cliboot.Config, cliboot.BootResult)
	agentCmd.GroupID = "ai"
	rootCmd.AddCommand(agentCmd)

	graphCmd := cligraph.NewGraphCmd(cliboot.Config)
	graphCmd.GroupID = "ai"
	rootCmd.AddCommand(graphCmd)

	a2aCmd := clia2a.NewA2ACmd(cliboot.Config)
	a2aCmd.GroupID = "ai"
	rootCmd.AddCommand(a2aCmd)

	learningCmd := clilearning.NewLearningCmd(cliboot.Config, cliboot.BootResult)
	learningCmd.GroupID = "ai"
	rootCmd.AddCommand(learningCmd)

	extensionCmd := cliextension.NewExtensionCmd(cliboot.Config)
	extensionCmd.GroupID = "ai"
	rootCmd.AddCommand(extensionCmd)

	librarianCmd := clilibrarian.NewLibrarianCmd(cliboot.Config, cliboot.BootResult)
	librarianCmd.GroupID = "ai"
	rootCmd.AddCommand(librarianCmd)

	metricsCmd := climetrics.NewMetricsCmd()
	metricsCmd.GroupID = "ai"
	rootCmd.AddCommand(metricsCmd)

	// --- Automation ---
	cronCmd := clicron.NewCronCmd(cliboot.BootResult)
	cronCmd.GroupID = "auto"
	rootCmd.AddCommand(cronCmd)

	workflowCmd := cliworkflow.NewWorkflowCmd(cliboot.BootResult)
	workflowCmd.GroupID = "auto"
	rootCmd.AddCommand(workflowCmd)

	runCmd := clirun.NewRunCmd(cliboot.BootResult)
	runCmd.GroupID = "auto"
	rootCmd.AddCommand(runCmd)

	provenanceCmd := cliprovenance.NewProvenanceCmd(cliboot.BootResult)
	provenanceCmd.GroupID = "auto"
	rootCmd.AddCommand(provenanceCmd)

	bgCmd := clibg.NewBgCmd(func() (*background.Manager, error) {
		return nil, fmt.Errorf("bg commands require a running server (use 'lango serve' first)")
	})
	bgCmd.GroupID = "auto"
	rootCmd.AddCommand(bgCmd)

	// --- Network & Economy ---
	p2pCmd := clip2p.NewP2PCmd(cliboot.BootResult)
	p2pCmd.GroupID = "net"
	rootCmd.AddCommand(p2pCmd)

	paymentCmd := clipayment.NewPaymentCmd(cliboot.BootResult)
	paymentCmd.GroupID = "net"
	rootCmd.AddCommand(paymentCmd)

	economyCmd := clieconomy.NewEconomyCmd(cliboot.Config)
	economyCmd.GroupID = "net"
	rootCmd.AddCommand(economyCmd)

	contractCmd := clicontract.NewContractCmd(cliboot.Config)
	contractCmd.GroupID = "net"
	rootCmd.AddCommand(contractCmd)

	accountCmd := cliaccount.NewAccountCmd(cliboot.BootResult)
	accountCmd.GroupID = "net"
	rootCmd.AddCommand(accountCmd)

	mcpCmd := climcp.NewMCPCmd(cliboot.Config, cliboot.BootResult)
	mcpCmd.GroupID = "net"
	rootCmd.AddCommand(mcpCmd)

	sandboxCmd := clisandbox.NewSandboxCmd(cliboot.Config, cliboot.BootResult)
	sandboxCmd.GroupID = "sys"
	rootCmd.AddCommand(sandboxCmd)

	alertsCmd := clialerts.NewAlertsCmd()
	alertsCmd.GroupID = "sys"
	rootCmd.AddCommand(alertsCmd)

	// --- Security & System (continued) ---
	approvalCmd := cliapproval.NewApprovalCmd(cliboot.Config)
	approvalCmd.GroupID = "sys"
	rootCmd.AddCommand(approvalCmd)

	healthCmd := healthCmd()
	healthCmd.GroupID = "sys"
	rootCmd.AddCommand(healthCmd)

	return rootCmd
}

func writeTUIStartupNotice(w io.Writer, logPath, line string) {
	fmt.Fprint(w, tui.Banner())
	fmt.Fprintf(w, "\n  Logs: %s\n", logPath)
	fmt.Fprintln(w, line)
}

func prepareTUIStartup(
	cfg *config.Config,
	profileName string,
	logFileName string,
	initFn func(logging.LogConfig) error,
	syncFn func() error,
	startupWriter io.Writer,
	line string,
) (func(), error) {
	logPath := filepath.Join(cfg.DataRoot, logFileName)
	if err := initFn(logging.LogConfig{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		OutputPath: logPath,
	}); err != nil {
		return nil, fmt.Errorf("init logging: %w", err)
	}

	// Redirect Go stdlib logger to the same file so third-party libraries
	// (e.g., ADK runner) that use log.Printf don't leak into the TUI.
	var logFile *os.File
	if logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		log.SetOutput(logFile)
	}

	tui.SetProfile(profileName)

	writeTUIStartupNotice(startupWriter, logPath, line)

	return func() {
		if logFile != nil {
			_ = logFile.Close()
		}
		_ = syncFn()
	}, nil
}

func runChat(initialMode string) error {
	boot, err := chatBootLoaderFn()
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer boot.Close()

	cfg := boot.Config
	if err := validateInitialMode(cfg, initialMode); err != nil {
		return err
	}
	cleanup, err := prepareTUIStartup(
		cfg,
		boot.ProfileName,
		"chat.log",
		chatLoggingInitFn,
		chatLoggingSyncFn,
		chatStartupErrWriter,
		"  Initializing...",
	)
	if err != nil {
		return err
	}
	defer cleanup()

	// Create app in local-chat mode (skip gateway/channels/automation lifecycle).
	application, err := chatAppBuilderFn(boot)
	if err != nil {
		return fmt.Errorf("create application: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := application.Start(ctx); err != nil {
		return fmt.Errorf("start application: %w", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = application.Stop(shutdownCtx)
	}()

	sessionKey := fmt.Sprintf("tui-%d", time.Now().UnixMilli())

	// Pre-create session and persist initial mode if --mode was provided
	// (mirrors runCockpit's mode handling).
	if initialMode != "" && application.Store != nil {
		s := &session.Session{Key: sessionKey}
		s.SetMode(initialMode)
		if err := application.Store.Create(s); err != nil {
			return fmt.Errorf("create initial session: %w", err)
		}
	}

	model := chat.New(chat.Deps{
		TurnRunner:   application.TurnRunner,
		Config:       cfg,
		SessionKey:   sessionKey,
		SessionStore: application.Store,
		EventBus:     application.EventBus,
	})

	// Hard session end: reads model.SessionKey() so /clear key changes
	// are respected (not the initial captured local).
	defer func() {
		if application.Store != nil {
			_ = application.Store.End(model.SessionKey())
		}
	}()

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	model.SetProgram(p)

	// Override TTY fallback with TUI approval provider.
	if composite, ok := application.ApprovalProvider.(*approval.CompositeProvider); ok {
		composite.SetTTYFallback(chat.NewTUIApprovalProvider(func(msg interface{}) {
			p.Send(msg)
		}))
	}

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI: %w", err)
	}

	return nil
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "serve",
		Short:   "Start the gateway server",
		GroupID: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := serveBootLoaderFn()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			cfg := boot.Config
			if err := serveLoggingInitFn(logging.LogConfig{
				Level:      cfg.Logging.Level,
				Format:     cfg.Logging.Format,
				OutputPath: cfg.Logging.OutputPath,
			}); err != nil {
				return fmt.Errorf("init logging: %w", err)
			}
			defer func() { _ = serveLoggingSyncFn() }()

			log := logging.Sugar()

			tui.SetProfile(boot.ProfileName)
			fmt.Fprint(cmd.OutOrStdout(), tui.ServeBanner())

			log.Infow("starting lango", "version", Version, "profile", boot.ProfileName)

			application, err := serveAppBuilderFn(boot)
			if err != nil {
				return fmt.Errorf("create application: %w", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigChan := make(chan os.Signal, 2)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			defer signal.Stop(sigChan)

			go watchServeSignals(ctx, application, log, sigChan, 10*time.Second, cancel, exitFn)

			if err := application.Start(ctx); err != nil {
				log.Errorw("startup error", "error", err)
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), startupSummary(cfg))

			serveAwaitShutdownFn(ctx)
			return nil
		},
	}
}

func watchServeSignals(
	ctx context.Context,
	application stoppableApplication,
	log *zap.SugaredLogger,
	sigChan <-chan os.Signal,
	shutdownTimeout time.Duration,
	cancel context.CancelFunc,
	forceExit func(int),
) {
	shutdownStarted := false

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-sigChan:
			if !ok {
				return
			}

			if shutdownStarted {
				log.Warn("received second interrupt, forcing exit")
				forceExit(130)
				return
			}

			shutdownStarted = true
			log.Info("shutting down...")
			go func() {
				shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
				defer shutdownCancel()
				if err := application.Stop(shutdownCtx); err != nil {
					log.Warnw("shutdown error", "error", err)
				}
				cancel()
			}()
		}
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print version information",
		GroupID: "start",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "lango %s (built %s)\n", Version, BuildTime)
		},
	}
}

func healthCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:           "health",
		Short:         "Check gateway health (replaces curl in Docker HEALTHCHECK)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			url := "http://localhost:" + strconv.Itoa(port) + "/health"
			client := newHealthHTTPClientFn()

			resp, err := client.Get(url)
			if err != nil {
				return fmt.Errorf("health check: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unhealthy: status %d", resp.StatusCode)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "ok")
			return nil
		},
	}

	cmd.Flags().IntVar(&port, "port", 18789, "gateway port to check")
	return cmd
}

func validateInitialMode(cfg *config.Config, initialMode string) error {
	if initialMode == "" {
		return nil
	}
	if _, ok := cfg.LookupMode(initialMode); !ok {
		return fmt.Errorf("unknown mode %q; valid modes can be listed via /mode", initialMode)
	}
	return nil
}

func configCmd() *cobra.Command {
	// Profile management subcommands (list, create, use, delete, import, export, validate).
	cmd := cliconfigcmd.NewConfigCmd(cliboot.BootResult)
	cmd.GroupID = "sys"

	// get/set/keys — config value inspection & modification.
	cmd.AddCommand(cliconfigcmd.NewGetCmd(cliboot.Config))
	var setBootResult *bootstrap.Result
	cmd.AddCommand(cliconfigcmd.NewSetCmd(
		func() (*config.Config, func(), error) {
			boot, err := cliboot.BootResult()
			if err != nil {
				return nil, nil, err
			}
			setBootResult = boot
			cleanup := func() {
				_ = boot.Close()
				setBootResult = nil
			}
			return boot.Config, cleanup, nil
		},
		func(cfg *config.Config) error {
			if setBootResult == nil {
				return fmt.Errorf("internal error: bootstrap result not available")
			}
			return setBootResult.Storage.ConfigProfiles().Save(
				context.Background(), setBootResult.ProfileName, cfg, nil,
			)
		},
	))
	cmd.AddCommand(cliconfigcmd.NewKeysCmd())

	return cmd
}

func startupSummary(cfg *config.Config) string {
	var channels []string
	if cfg.Channels.Telegram.Enabled {
		channels = append(channels, "telegram")
	}
	if cfg.Channels.Discord.Enabled {
		channels = append(channels, "discord")
	}
	if cfg.Channels.Slack.Enabled {
		channels = append(channels, "slack")
	}

	channelDetail := "none"
	if len(channels) > 0 {
		channelDetail = strings.Join(channels, ", ")
	}

	features := []tui.FeatureLine{
		{Name: "Gateway", Enabled: cfg.Server.HTTPEnabled, Detail: fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)},
		{Name: "Channels", Enabled: len(channels) > 0, Detail: channelDetail},
		{Name: "Knowledge", Enabled: cfg.Knowledge.Enabled},
		{Name: "Embedding & RAG", Enabled: cfg.Embedding.Provider != "", Detail: cfg.Embedding.Provider},
		{Name: "Graph", Enabled: cfg.Graph.Enabled},
		{Name: "Obs. Memory", Enabled: cfg.ObservationalMemory.Enabled},
		{Name: "Cron", Enabled: cfg.Cron.Enabled},
		{Name: "MCP", Enabled: cfg.MCP.Enabled, Detail: mcpServerCount(cfg)},
		{Name: "P2P", Enabled: cfg.P2P.Enabled},
		{Name: "Payment", Enabled: cfg.Payment.Enabled},
		{Name: "Provenance", Enabled: cfg.Provenance.Enabled},
	}

	return tui.StartupSummary(features)
}

func mcpServerCount(cfg *config.Config) string {
	if !cfg.MCP.Enabled {
		return ""
	}
	n := len(cfg.MCP.Servers)
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d server(s)", n)
}

// withChannels controls whether cockpit starts live channel adapters.
// Default false to avoid conflicts with a running `lango serve`.
// Set via --with-channels flag on `lango cockpit`.
var withChannels bool

func cockpitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cockpit",
		Short:   "Launch multi-panel operator dashboard",
		GroupID: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isInteractiveFn() {
				return fmt.Errorf("cockpit requires an interactive terminal")
			}
			modeName, _ := cmd.Flags().GetString("mode")
			return runCockpitFn(modeName)
		},
	}
	cmd.Flags().String("mode", "", "Initial session mode (e.g., code-review, research, debug)")
	cmd.Flags().BoolVar(&withChannels, "with-channels", false,
		"Start live channel adapters (Telegram/Discord/Slack). "+
			"Only use when no lango serve is running with the same credentials.")
	return cmd
}

func chatCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "chat",
		Short:   "Launch plain chat TUI",
		GroupID: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isInteractiveFn() {
				return fmt.Errorf("chat requires an interactive terminal")
			}
			modeName, _ := cmd.Flags().GetString("mode")
			return runChatFn(modeName)
		},
	}
}

func buildMissionControlDeps(
	application *app.App,
	cfg *config.Config,
	workDir string,
	sessionKey string,
	store session.Store,
	profileName string,
	configStore storage.ConfigProfileStore,
	pendingApprovals *cockpit.PendingApprovalRegistry,
	learningBuffer *cockpit.LearningSuggestionBuffer,
	activityBuffer *cockpit.MissionActivityBuffer,
) cockpit.Deps {
	return cockpit.Deps{
		TurnRunner:         application.TurnRunner,
		Config:             cfg,
		WorkDir:            workDir,
		SessionKey:         sessionKey,
		SessionStore:       store,
		ToolCatalog:        application.ToolCatalog,
		MetricsCollector:   application.MetricsCollector,
		ConfigStore:        configStore,
		ProfileName:        profileName,
		BackgroundManager:  application.BackgroundManager,
		EventBus:           application.EventBus,
		ApprovalHistory:    application.ApprovalHistory,
		GrantStore:         application.GrantStore,
		MissionReader:      application.MissionStore,
		ProposalReader:     application.ProposalRegistry,
		ProposalService:    application.ProposalService,
		LoopInquiryReader:  application.LoopInquiryReader,
		LoopDeadReader:     application.LoopDeadLetterReader,
		LoopCronReader:     application.LoopCronReader,
		CollabMissionLinks: application.CollaborationMissionLinkReader,
		CollabAgentRuns:    application.CollaborationAgentRunReader,
		CollabDelegations:  application.CollaborationDelegationReader,
		CollabRuntime:      application.CollaborationRuntimeReader,
		MissionService:     application.MissionService,
		PendingApprovals:   pendingApprovals,
		LearningBuffer:     learningBuffer,
		ActivityBuffer:     activityBuffer,
		RunLedgerStore:     application.RunLedgerStore,
		AgentRunStore:      application.AgentRunStore,
	}
}

func runCockpit(initialMode string) error {
	boot, err := cockpitBootLoaderFn()
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer boot.Close()

	cfg := boot.Config
	if err := validateInitialMode(cfg, initialMode); err != nil {
		return err
	}
	cleanup, err := prepareTUIStartup(
		cfg,
		boot.ProfileName,
		"cockpit.log",
		cockpitLoggingInitFn,
		cockpitLoggingSyncFn,
		cockpitStartupErrWriter,
		"  Initializing cockpit...",
	)
	if err != nil {
		return err
	}
	defer cleanup()

	// Use Cockpit mode (channels enabled) only when explicitly requested
	// via --with-channels to avoid conflict with a running `lango serve`.
	appMode := app.WithLocalChat()
	if withChannels {
		appMode = app.WithCockpit()
	}
	application, err := cockpitAppBuilderFn(boot, appMode)
	if err != nil {
		return fmt.Errorf("create application: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := application.Start(ctx); err != nil {
		return fmt.Errorf("start application: %w", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(
			context.Background(), 10*time.Second,
		)
		defer shutdownCancel()
		_ = application.Stop(shutdownCtx)
	}()

	// Create channel tracker for cockpit status display.
	tracker := cockpit.NewChannelTracker(application.EventBus)

	// Seed tracker with known channel names (Start status updated later).
	for _, ch := range application.Channels {
		tracker.SeedChannel(ch.Name(), false)
	}

	// Channel shutdown: cancel ctx first so in-flight requests unblock,
	// then stop channels. Using a single defer to guarantee ordering
	// (cancel before stop) and avoid LIFO defer reversal.
	defer func() {
		cancel() // unblock any in-flight channel workers waiting on ctx
		for _, ch := range application.Channels {
			stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = ch.Stop(stopCtx)
			stopCancel()
		}
	}()

	sessionKey := fmt.Sprintf("cockpit-%d", time.Now().UnixMilli())

	// Hard session end (TUI quit): bounded best-effort summarize/index.
	// The EntStore processor honors its own hardEndTimeout so this call
	// returns within a short bound even on a slow summarizer. Errors are
	// ignored — a missing session (if the first turn never ran) is fine.
	defer func() {
		if application.Store != nil {
			_ = application.Store.End(sessionKey)
		}
	}()

	// Pre-create the session and persist initial mode if --mode was provided.
	if initialMode != "" {
		s := &session.Session{Key: sessionKey}
		s.SetMode(initialMode)
		if err := application.Store.Create(s); err != nil {
			return fmt.Errorf("create initial session: %w", err)
		}
	}

	pendingApprovals := cockpit.NewPendingApprovalRegistry()
	learningBuffer := cockpit.NewLearningSuggestionBuffer(nil)
	activityBuffer := cockpit.NewMissionActivityBuffer()

	cockpitDeps := buildMissionControlDeps(
		application,
		cfg,
		currentWorkDir(cfg),
		sessionKey,
		application.Store,
		boot.ProfileName,
		boot.Storage.ConfigProfiles(),
		pendingApprovals,
		learningBuffer,
		activityBuffer,
	)

	model := cockpit.New(cockpitDeps)
	missionComposer := model.ChatModel()
	if missionComposer == nil {
		return fmt.Errorf("cockpit chat model is not available")
	}

	registerCockpitPages(
		model,
		application,
		cfg,
		boot.ProfileName,
		boot.Storage.ConfigProfiles(),
		cockpitDeps,
		missionComposer,
	)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	model.SetProgram(p)
	model.SetChannelTracker(tracker)

	// Wire runtime tracker for live token/delegation/recovery metrics.
	runtimeTracker := cockpit.NewRuntimeTracker(application.EventBus, p, sessionKey)
	model.SetRuntimeTracker(runtimeTracker)

	// Wire cockpit-lifetime shared state subscriptions before runtime starts
	// emitting session events.
	cockpit.SubscribeMissionControlEvents(application.EventBus, sessionKey, learningBuffer, activityBuffer)

	// Wire channel events from EventBus to TUI — BEFORE starting channels
	// so no early inbound messages are dropped (EventBus drops unhandled events).
	cockpit.SubscribeChannelEvents(application.EventBus, p)

	// Start channel polling/socket loops AFTER subscribe is wired.
	for _, ch := range application.Channels {
		ch := ch
		go func() {
			err := ch.Start(ctx)
			tracker.SeedChannel(ch.Name(), err == nil)
			if err != nil {
				fmt.Fprintf(cockpitStartupErrWriter, "channel start (%s): %v\n", ch.Name(), err)
			}
		}()
	}

	if composite, ok := application.ApprovalProvider.(*approval.CompositeProvider); ok {
		composite.SetTTYFallback(chat.NewTUIApprovalProvider(func(msg interface{}) {
			p.Send(msg)
		}))
	}

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI: %w", err)
	}

	return nil
}

func registerCockpitPages(
	model *cockpit.Model,
	application *app.App,
	cfg *config.Config,
	profileName string,
	configStore storage.ConfigProfileStore,
	cockpitDeps cockpit.Deps,
	missionComposer *chat.ChatModel,
) {
	model.RegisterPage(cockpit.PageMissionControl,
		pages.NewMissionControlPage(cockpitDeps, missionComposer))
	model.RegisterPage(cockpit.PageTools,
		pages.NewToolsPage(application.ToolCatalog))

	var statusProvider func() []types.FeatureStatus
	if application.FeatureStatuses != nil {
		statusProvider = application.FeatureStatuses.All
	}
	model.RegisterPage(cockpit.PageStatus,
		pages.NewStatusPage(statusProvider, application.MetricsCollector, cfg))

	model.RegisterPage(cockpit.PageSettings,
		pages.NewSettingsPage(cfg, configStore, profileName))

	model.RegisterPage(cockpit.PageSessions,
		pages.NewSessionsPage(func(ctx context.Context) ([]session.SessionSummary, error) {
			return application.Store.ListSessions(ctx)
		}))
	if application.BackgroundManager != nil {
		var actioner pages.TaskActioner = &bgTaskActioner{mgr: application.BackgroundManager}
		model.RegisterPage(cockpit.PageTasks,
			pages.NewTasksPage(&bgTaskLister{mgr: application.BackgroundManager}, actioner))
	} else {
		model.RegisterPage(cockpit.PageTasks, pages.NewTasksPage(nil, nil))
	}
	model.RegisterPage(cockpit.PageDeadLetters, pages.NewDeadLettersPage(nil, nil))
	if deadLetterBridge := cockpit.NewDeadLetterToolBridge(application.ToolCatalog); deadLetterBridge.Ready() {
		var retryFn pages.DeadLetterRetryFn
		if deadLetterBridge.CanRetry() {
			retryFn = deadLetterBridge.Retry
		}
		listFn := func(ctx context.Context, opts pages.DeadLetterListOptions) ([]postadjudicationstatus.DeadLetterBacklogEntry, error) {
			return deadLetterBridge.List(ctx, cockpitDeadLetterListOptions(opts))
		}
		model.RegisterPage(cockpit.PageDeadLetters,
			pages.NewDeadLettersPage(listFn, deadLetterBridge.Detail, retryFn))
	}
	model.RegisterPage(cockpit.PageApprovals,
		pages.NewApprovalsPage(application.ApprovalHistory, application.GrantStore))
}

func runWorkbench(initialMode string) error {
	boot, err := workbenchBootLoaderFn()
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer boot.Close()

	cfg := boot.Config
	if err := validateInitialMode(cfg, initialMode); err != nil {
		return err
	}
	cleanup, err := prepareTUIStartup(
		cfg,
		boot.ProfileName,
		"workbench.log",
		workbenchLoggingInitFn,
		workbenchLoggingSyncFn,
		workbenchStartupErrWriter,
		"  Initializing workbench...",
	)
	if err != nil {
		return err
	}
	defer cleanup()

	application, err := workbenchAppBuilderFn(boot)
	if err != nil {
		return fmt.Errorf("create application: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := application.Start(ctx); err != nil {
		return fmt.Errorf("start application: %w", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = application.Stop(shutdownCtx)
	}()

	sessionKey := fmt.Sprintf("workbench-%d", time.Now().UnixMilli())
	defer func() {
		if application.Store != nil {
			_ = application.Store.End(sessionKey)
		}
	}()

	if initialMode != "" {
		s := &session.Session{Key: sessionKey}
		s.SetMode(initialMode)
		if err := application.Store.Create(s); err != nil {
			return fmt.Errorf("create initial session: %w", err)
		}
	}

	pendingApprovals := cockpit.NewPendingApprovalRegistry()
	learningBuffer := cockpit.NewLearningSuggestionBuffer(nil)
	activityBuffer := cockpit.NewMissionActivityBuffer()

	deps := buildMissionControlDeps(
		application,
		cfg,
		currentWorkDir(cfg),
		sessionKey,
		application.Store,
		boot.ProfileName,
		boot.Storage.ConfigProfiles(),
		pendingApprovals,
		learningBuffer,
		activityBuffer,
	)

	model := workbench.New(deps)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	model.SetProgram(p)
	model.SetRuntimeTracker(cockpit.NewRuntimeTracker(application.EventBus, p, sessionKey))

	if composite, ok := application.ApprovalProvider.(*approval.CompositeProvider); ok {
		composite.SetTTYFallback(chat.NewTUIApprovalProvider(func(msg interface{}) {
			p.Send(msg)
		}))
	}

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI: %w", err)
	}

	return nil
}

func currentWorkDir(cfg *config.Config) string {
	if cfg != nil {
		if workDir := strings.TrimSpace(cfg.Tools.Exec.WorkDir); workDir != "" {
			return workDir
		}
	}
	workDir, _ := os.Getwd()
	return strings.TrimSpace(workDir)
}

func cockpitDeadLetterListOptions(opts pages.DeadLetterListOptions) cockpit.DeadLetterListOptions {
	return cockpit.DeadLetterListOptions{
		Query:                     opts.Query,
		Adjudication:              opts.Adjudication,
		LatestStatusSubtype:       opts.LatestStatusSubtype,
		LatestStatusSubtypeFamily: opts.LatestStatusSubtypeFamily,
		AnyMatchFamily:            opts.AnyMatchFamily,
		ManualReplayActor:         opts.ManualReplayActor,
		DeadLetteredAfter:         opts.DeadLetteredAfter,
		DeadLetteredBefore:        opts.DeadLetteredBefore,
		DeadLetterReasonQuery:     opts.DeadLetterReasonQuery,
		LatestDispatchReference:   opts.LatestDispatchReference,
	}
}

// bgTaskLister adapts background.Manager to pages.TaskLister.
type bgTaskLister struct {
	mgr *background.Manager
}

func (b *bgTaskLister) ListTasks() []pages.TaskInfo {
	snapshots := b.mgr.List()

	// Sort by StartedAt descending for stable ordering.
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].StartedAt.After(snapshots[j].StartedAt)
	})

	tasks := make([]pages.TaskInfo, len(snapshots))
	for i, s := range snapshots {
		tasks[i] = pages.TaskInfo{
			ID:            s.ID,
			Prompt:        s.Prompt,
			Status:        s.StatusText,
			Elapsed:       taskElapsed(s),
			Result:        s.Result,
			Error:         s.Error,
			OriginChannel: s.OriginChannel,
			TokensUsed:    s.TokensUsed,
		}
	}
	return tasks
}

// bgTaskActioner adapts background.Manager to pages.TaskActioner.
type bgTaskActioner struct {
	mgr *background.Manager
}

func (b *bgTaskActioner) CancelTask(id string) error {
	return b.mgr.Cancel(id)
}

func (b *bgTaskActioner) RetryTask(ctx context.Context, id string) error {
	snap, err := b.mgr.Status(id)
	if err != nil {
		return fmt.Errorf("retry task %s: %w", id, err)
	}
	_, err = b.mgr.Submit(ctx, snap.Prompt, background.Origin{
		Channel: snap.OriginChannel,
		Session: snap.OriginSession,
	})
	if err != nil {
		return fmt.Errorf("retry task %s: %w", id, err)
	}
	return nil
}

// taskElapsed computes the correct elapsed duration for a task snapshot.
func taskElapsed(s background.TaskSnapshot) time.Duration {
	if s.StartedAt.IsZero() {
		return 0 // pending, not yet started
	}
	if !s.CompletedAt.IsZero() {
		return s.CompletedAt.Sub(s.StartedAt) // terminal: freeze at actual runtime
	}
	return time.Since(s.StartedAt) // running: wall-clock
}
