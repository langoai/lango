package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
	clia2a "github.com/langoai/lango/internal/cli/a2a"
	cliagent "github.com/langoai/lango/internal/cli/agent"
	cliapproval "github.com/langoai/lango/internal/cli/approval"
	cliconfigcmd "github.com/langoai/lango/internal/cli/configcmd"
	clibg "github.com/langoai/lango/internal/cli/bg"
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
	clisecurity "github.com/langoai/lango/internal/cli/security"
	"github.com/langoai/lango/internal/cli/settings"
	cliaccount "github.com/langoai/lango/internal/cli/smartaccount"
	clistatus "github.com/langoai/lango/internal/cli/status"
	"github.com/langoai/lango/internal/cli/tui"
	cliworkflow "github.com/langoai/lango/internal/cli/workflow"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/sandbox"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Check if running as sandbox worker subprocess.
	// Worker mode is used for process-isolated tool execution in P2P.
	if sandbox.IsWorkerMode() {
		// Phase 1: no tools registered in worker — the subprocess executor
		// is wired at the application level. This early exit prevents the
		// worker from initializing cobra and the full application stack.
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

	statusCmd := clistatus.NewStatusCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	statusCmd.GroupID = "start"
	rootCmd.AddCommand(statusCmd)

	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(configCmd())

	// --- Security & System ---
	securityCmd := clisecurity.NewSecurityCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	securityCmd.GroupID = "sys"
	rootCmd.AddCommand(securityCmd)

	// --- AI & Knowledge ---
	memoryCmd := climemory.NewMemoryCmd(func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	})
	memoryCmd.GroupID = "ai"
	rootCmd.AddCommand(memoryCmd)

	agentCmd := cliagent.NewAgentCmd(func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	})
	agentCmd.GroupID = "ai"
	rootCmd.AddCommand(agentCmd)

	graphCmd := cligraph.NewGraphCmd(func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	})
	graphCmd.GroupID = "ai"
	rootCmd.AddCommand(graphCmd)

	a2aCmd := clia2a.NewA2ACmd(func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	})
	a2aCmd.GroupID = "ai"
	rootCmd.AddCommand(a2aCmd)

	learningCfgLoader := func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	}
	learningBootLoader := func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	}
	learningCmd := clilearning.NewLearningCmd(learningCfgLoader, learningBootLoader)
	learningCmd.GroupID = "ai"
	rootCmd.AddCommand(learningCmd)

	librarianCfgLoader := func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	}
	librarianBootLoader := func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	}
	librarianCmd := clilibrarian.NewLibrarianCmd(librarianCfgLoader, librarianBootLoader)
	librarianCmd.GroupID = "ai"
	rootCmd.AddCommand(librarianCmd)

	metricsCmd := climetrics.NewMetricsCmd()
	metricsCmd.GroupID = "ai"
	rootCmd.AddCommand(metricsCmd)

	// --- Automation ---
	cronCmd := clicron.NewCronCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	cronCmd.GroupID = "auto"
	rootCmd.AddCommand(cronCmd)

	workflowCmd := cliworkflow.NewWorkflowCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	workflowCmd.GroupID = "auto"
	rootCmd.AddCommand(workflowCmd)

	bgCmd := clibg.NewBgCmd(func() (*background.Manager, error) {
		return nil, fmt.Errorf("bg commands require a running server (use 'lango serve' first)")
	})
	bgCmd.GroupID = "auto"
	rootCmd.AddCommand(bgCmd)

	// --- Network & Economy ---
	p2pCmd := clip2p.NewP2PCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	p2pCmd.GroupID = "net"
	rootCmd.AddCommand(p2pCmd)

	paymentCmd := clipayment.NewPaymentCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	paymentCmd.GroupID = "net"
	rootCmd.AddCommand(paymentCmd)

	economyCfgLoader := func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	}
	economyCmd := clieconomy.NewEconomyCmd(economyCfgLoader)
	economyCmd.GroupID = "net"
	rootCmd.AddCommand(economyCmd)

	contractCfgLoader := func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	}
	contractCmd := clicontract.NewContractCmd(contractCfgLoader)
	contractCmd.GroupID = "net"
	rootCmd.AddCommand(contractCmd)

	accountCmd := cliaccount.NewAccountCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	accountCmd.GroupID = "net"
	rootCmd.AddCommand(accountCmd)

	mcpCfgLoader := func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	}
	mcpBootLoader := func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	}
	mcpCmd := climcp.NewMCPCmd(mcpCfgLoader, mcpBootLoader)
	mcpCmd.GroupID = "net"
	rootCmd.AddCommand(mcpCmd)

	// --- Security & System (continued) ---
	approvalCmd := cliapproval.NewApprovalCmd(func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	})
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

// bootstrapForConfig creates a bootstrap result for config subcommands.
func bootstrapForConfig() (*bootstrap.Result, error) {
	return bootstrap.Run(bootstrap.Options{})
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "serve",
		Short:   "Start the gateway server",
		GroupID: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Bootstrap: DB + crypto + config profile
			boot, err := bootstrap.Run(bootstrap.Options{})
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			// Initialize logging
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

			// Print serve banner before starting
			tui.SetProfile(boot.ProfileName)
			fmt.Print(tui.ServeBanner())

			log.Infow("starting lango", "version", Version, "profile", boot.ProfileName)

			// Create application
			application, err := app.New(boot)
			if err != nil {
				return fmt.Errorf("create application: %w", err)
			}

			// Setup graceful shutdown
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

			// Start application
			if err := application.Start(ctx); err != nil {
				log.Errorw("startup error", "error", err)
				return err
			}

			// Print startup summary
			fmt.Print(startupSummary(cfg))

			// Wait for shutdown
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
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Configuration profile management",
		GroupID: "sys",
		Long: `Configuration profile management.

Manage multiple configuration profiles for different environments or setups.

See Also:
  lango settings  - Interactive settings editor (TUI)
  lango onboard   - Guided setup wizard
  lango doctor    - Diagnose configuration issues`,
	}

	cmd.AddCommand(configListCmd())
	cmd.AddCommand(configCreateCmd())
	cmd.AddCommand(configUseCmd())
	cmd.AddCommand(configDeleteCmd())
	cmd.AddCommand(configImportCmd())
	cmd.AddCommand(configExportCmd())
	cmd.AddCommand(configValidateCmd())

	// get/set/keys — config value inspection & modification
	cmd.AddCommand(cliconfigcmd.NewGetCmd(func() (*config.Config, error) {
		boot, err := bootstrapForConfig()
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	}))
	var setBootResult *bootstrap.Result
	cmd.AddCommand(cliconfigcmd.NewSetCmd(
		func() (*config.Config, func(), error) {
			boot, err := bootstrapForConfig()
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

func configListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootstrapForConfig()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			profiles, err := boot.ConfigStore.List(context.Background())
			if err != nil {
				return fmt.Errorf("list profiles: %w", err)
			}

			if len(profiles) == 0 {
				fmt.Println("No profiles found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tACTIVE\tVERSION\tCREATED\tUPDATED")
			for _, p := range profiles {
				active := ""
				if p.Active {
					active = "*"
				}
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
					p.Name,
					active,
					p.Version,
					p.CreatedAt.Format(time.DateTime),
					p.UpdatedAt.Format(time.DateTime),
				)
			}
			return w.Flush()
		},
	}
}

func configCreateCmd() *cobra.Command {
	var preset string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new profile (optionally from a preset)",
		Long: `Create a new configuration profile.

Use --preset to start from a purpose-built template:
  minimal       Basic AI agent (quick start)
  researcher    Knowledge, RAG, Graph (research/analysis)
  collaborator  P2P team, payment, workspace (collaboration)
  full          All features enabled (power user)

Examples:
  lango config create my-profile
  lango config create research-bot --preset researcher`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if preset != "" && !config.IsValidPreset(preset) {
				return fmt.Errorf("unknown preset %q (valid: minimal, researcher, collaborator, full)", preset)
			}

			boot, err := bootstrapForConfig()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			ctx := context.Background()

			exists, err := boot.ConfigStore.Exists(ctx, name)
			if err != nil {
				return fmt.Errorf("check profile: %w", err)
			}
			if exists {
				return fmt.Errorf("profile %q already exists", name)
			}

			var cfg *config.Config
			if preset != "" {
				cfg = config.PresetConfig(preset)
			} else {
				cfg = config.DefaultConfig()
			}

			if err := boot.ConfigStore.Save(ctx, name, cfg); err != nil {
				return fmt.Errorf("create profile: %w", err)
			}

			if preset != "" {
				fmt.Printf("Profile %q created from preset %q.\n", name, preset)
			} else {
				fmt.Printf("Profile %q created with default configuration.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&preset, "preset", "", "Preset template (minimal, researcher, collaborator, full)")
	return cmd
}

func configUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Switch to a different configuration profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			boot, err := bootstrapForConfig()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			if err := boot.ConfigStore.SetActive(context.Background(), name); err != nil {
				return fmt.Errorf("switch profile: %w", err)
			}

			fmt.Printf("Switched to profile %q.\n", name)
			return nil
		},
	}
}

func configDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a configuration profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !force {
				fmt.Printf("Delete profile %q? This cannot be undone. [y/N]: ", name)
				var answer string
				_, _ = fmt.Scanln(&answer)
				if answer != "y" && answer != "Y" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			boot, err := bootstrapForConfig()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			if err := boot.ConfigStore.Delete(context.Background(), name); err != nil {
				return fmt.Errorf("delete profile: %w", err)
			}

			fmt.Printf("Profile %q deleted.\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	return cmd
}

func configImportCmd() *cobra.Command {
	var profileName string

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import and encrypt a JSON config (source file is deleted after import)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			boot, err := bootstrapForConfig()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			ctx := context.Background()
			if err := configstore.MigrateFromJSON(ctx, boot.ConfigStore, filePath, profileName); err != nil {
				return fmt.Errorf("import config: %w", err)
			}

			fmt.Printf("Imported %q as profile %q (now active).\n", filePath, profileName)
			fmt.Println("Source file deleted for security.")
			return nil
		},
	}

	cmd.Flags().StringVar(&profileName, "profile", "default", "name for the imported profile")
	return cmd
}

func configExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export <name>",
		Short: "Export a profile as plaintext JSON (requires passphrase verification)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Verify passphrase before export.
			// Bootstrap already validates the passphrase, so reaching here
			// means the passphrase is correct.
			boot, err := bootstrapForConfig()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			cfg, err := boot.ConfigStore.Load(context.Background(), name)
			if err != nil {
				return fmt.Errorf("load profile: %w", err)
			}

			fmt.Fprintln(os.Stderr, "WARNING: exported configuration contains sensitive values in plaintext.")

			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}

			fmt.Println(string(data))
			return nil
		},
	}
}

func configValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the active configuration profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootstrapForConfig()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			if err := config.Validate(boot.Config); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			fmt.Printf("Profile %q configuration is valid.\n", boot.ProfileName)
			return nil
		},
	}
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
		{"Gateway", cfg.Server.HTTPEnabled, fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)},
		{"Channels", len(channels) > 0, channelDetail},
		{"Knowledge", cfg.Knowledge.Enabled, ""},
		{"Embedding & RAG", cfg.Embedding.Provider != "", cfg.Embedding.Provider},
		{"Graph", cfg.Graph.Enabled, ""},
		{"Obs. Memory", cfg.ObservationalMemory.Enabled, ""},
		{"Cron", cfg.Cron.Enabled, ""},
		{"MCP", cfg.MCP.Enabled, mcpServerCount(cfg)},
		{"P2P", cfg.P2P.Enabled, ""},
		{"Payment", cfg.Payment.Enabled, ""},
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
