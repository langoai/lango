package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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
	cliapproval "github.com/langoai/lango/internal/cli/approval"
	clibg "github.com/langoai/lango/internal/cli/bg"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/cli/cliboot"
	cliconfigcmd "github.com/langoai/lango/internal/cli/configcmd"
	clicontract "github.com/langoai/lango/internal/cli/contract"
	clicron "github.com/langoai/lango/internal/cli/cron"
	"github.com/langoai/lango/internal/cli/doctor"
	clieconomy "github.com/langoai/lango/internal/cli/economy"
	cligraph "github.com/langoai/lango/internal/cli/graph"
	clilearning "github.com/langoai/lango/internal/cli/learning"
	clilibrarian "github.com/langoai/lango/internal/cli/librarian"
	climcp "github.com/langoai/lango/internal/cli/mcp"
	climemory "github.com/langoai/lango/internal/cli/memory"
	climetrics "github.com/langoai/lango/internal/cli/metrics"
	"github.com/langoai/lango/internal/cli/onboard"
	clip2p "github.com/langoai/lango/internal/cli/p2p"
	clipayment "github.com/langoai/lango/internal/cli/payment"
	clisandbox "github.com/langoai/lango/internal/cli/sandbox"
	cliprovenance "github.com/langoai/lango/internal/cli/provenance"
	clirun "github.com/langoai/lango/internal/cli/run"
	clisecurity "github.com/langoai/lango/internal/cli/security"
	"github.com/langoai/lango/internal/cli/settings"
	cliaccount "github.com/langoai/lango/internal/cli/smartaccount"
	clistatus "github.com/langoai/lango/internal/cli/status"
	"github.com/langoai/lango/internal/cli/tui"
	cliworkflow "github.com/langoai/lango/internal/cli/workflow"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/sandbox"
	"go.uber.org/zap"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

var exitFn = os.Exit

type stoppableApplication interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func main() {
	// Check if running as sandbox worker subprocess.
	if sandbox.IsWorkerMode() {
		sandbox.RunWorker(sandbox.ToolRegistry{})
		return
	}

	tui.SetVersionInfo(Version, BuildTime)

	rootCmd := &cobra.Command{
		Use:   "lango",
		Short: "Lango - Fast AI Agent in Go",
		Long:  `Lango is a high-performance AI agent built with Go, supporting multiple channels and tools.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChat()
		},
	}

	rootCmd.AddGroup(
		&cobra.Group{ID: "start", Title: "Getting Started:"},
		&cobra.Group{ID: "ai", Title: "AI & Knowledge:"},
		&cobra.Group{ID: "auto", Title: "Automation:"},
		&cobra.Group{ID: "net", Title: "Network & Economy:"},
		&cobra.Group{ID: "sys", Title: "Security & System:"},
	)

	// --- Getting Started ---
	rootCmd.AddCommand(serveCmd())

	onboardCmd := onboard.NewCommand()
	onboardCmd.GroupID = "start"
	rootCmd.AddCommand(onboardCmd)

	doctorCmd := doctor.NewCommand()
	doctorCmd.GroupID = "start"
	rootCmd.AddCommand(doctorCmd)

	settingsCmd := settings.NewCommand()
	settingsCmd.GroupID = "start"
	rootCmd.AddCommand(settingsCmd)

	statusCmd := clistatus.NewStatusCmd(cliboot.BootResult)
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

	sandboxCmd := clisandbox.NewSandboxCmd(cliboot.Config)
	sandboxCmd.GroupID = "sys"
	rootCmd.AddCommand(sandboxCmd)

	// --- Security & System (continued) ---
	approvalCmd := cliapproval.NewApprovalCmd(cliboot.Config)
	approvalCmd.GroupID = "sys"
	rootCmd.AddCommand(approvalCmd)

	healthCmd := healthCmd()
	healthCmd.GroupID = "sys"
	rootCmd.AddCommand(healthCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runChat() error {
	boot, err := cliboot.BootResult()
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer boot.DBClient.Close()

	cfg := boot.Config
	// TUI mode: redirect logging to file (stderr output corrupts alt-screen TUI).
	logPath := filepath.Join(cfg.DataRoot, "chat.log")
	if err := logging.Init(logging.LogConfig{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		OutputPath: logPath,
	}); err != nil {
		return fmt.Errorf("init logging: %w", err)
	}
	defer func() { _ = logging.Sync() }()

	// Redirect Go stdlib logger to the same file so third-party libraries
	// (e.g., ADK runner) that use log.Printf don't leak into the TUI.
	if logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	tui.SetProfile(boot.ProfileName)

	fmt.Fprint(os.Stderr, tui.Banner())
	fmt.Fprintf(os.Stderr, "\n  Logs: %s\n", logPath)
	fmt.Fprintln(os.Stderr, "  Initializing...")

	// Create app in local-chat mode (skip gateway/channels/automation lifecycle).
	application, err := app.New(boot, app.WithLocalChat())
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

	model := chat.New(chat.Deps{
		TurnRunner: application.TurnRunner,
		Config:     cfg,
		SessionKey: sessionKey,
	})

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
			boot, err := cliboot.BootResult()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			cfg := boot.Config
			if err := logging.Init(logging.LogConfig{
				Level:      cfg.Logging.Level,
				Format:     cfg.Logging.Format,
				OutputPath: cfg.Logging.OutputPath,
			}); err != nil {
				return fmt.Errorf("init logging: %w", err)
			}
			defer func() { _ = logging.Sync() }()

			log := logging.Sugar()

			tui.SetProfile(boot.ProfileName)
			fmt.Print(tui.ServeBanner())

			log.Infow("starting lango", "version", Version, "profile", boot.ProfileName)

			application, err := app.New(boot)
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

			fmt.Print(startupSummary(cfg))

			<-ctx.Done()
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
			fmt.Printf("lango %s (built %s)\n", Version, BuildTime)
		},
	}
}

func healthCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check gateway health (replaces curl in Docker HEALTHCHECK)",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := "http://localhost:" + strconv.Itoa(port) + "/health"
			client := &http.Client{Timeout: 5 * time.Second}

			resp, err := client.Get(url)
			if err != nil {
				return fmt.Errorf("health check: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unhealthy: status %d", resp.StatusCode)
			}

			fmt.Println("ok")
			return nil
		},
	}

	cmd.Flags().IntVar(&port, "port", 18789, "gateway port to check")
	return cmd
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
				boot.DBClient.Close()
				setBootResult = nil
			}
			return boot.Config, cleanup, nil
		},
		func(cfg *config.Config) error {
			if setBootResult == nil {
				return fmt.Errorf("internal error: bootstrap result not available")
			}
			return setBootResult.ConfigStore.Save(
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
