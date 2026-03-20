package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
	clia2a "github.com/langoai/lango/internal/cli/a2a"
	cliagent "github.com/langoai/lango/internal/cli/agent"
	cliapproval "github.com/langoai/lango/internal/cli/approval"
	clibg "github.com/langoai/lango/internal/cli/bg"
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
	clirun "github.com/langoai/lango/internal/cli/run"
	climemory "github.com/langoai/lango/internal/cli/memory"
	climetrics "github.com/langoai/lango/internal/cli/metrics"
	"github.com/langoai/lango/internal/cli/onboard"
	clip2p "github.com/langoai/lango/internal/cli/p2p"
	clipayment "github.com/langoai/lango/internal/cli/payment"
	clisecurity "github.com/langoai/lango/internal/cli/security"
	"github.com/langoai/lango/internal/cli/settings"
	cliaccount "github.com/langoai/lango/internal/cli/smartaccount"
	clistatus "github.com/langoai/lango/internal/cli/status"
	"github.com/langoai/lango/internal/cli/tui"
	cliworkflow "github.com/langoai/lango/internal/cli/workflow"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/sandbox"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

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

	agentCmd := cliagent.NewAgentCmd(cliboot.Config)
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

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				log.Info("shutting down...")
				shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
				defer shutdownCancel()
				if err := application.Stop(shutdownCtx); err != nil {
					log.Warnw("shutdown error", "error", err)
				}
				cancel()
			}()

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
				context.Background(), setBootResult.ProfileName, cfg,
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
