package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	cliagent "github.com/langoai/lango/internal/cli/agent"
	clicron "github.com/langoai/lango/internal/cli/cron"
	"github.com/langoai/lango/internal/cli/doctor"
	cligraph "github.com/langoai/lango/internal/cli/graph"
	climemory "github.com/langoai/lango/internal/cli/memory"
	"github.com/langoai/lango/internal/cli/onboard"
	clip2p "github.com/langoai/lango/internal/cli/p2p"
	"github.com/langoai/lango/internal/cli/tui"
	clipayment "github.com/langoai/lango/internal/cli/payment"
	clisecurity "github.com/langoai/lango/internal/cli/security"
	"github.com/langoai/lango/internal/cli/settings"
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
		&cobra.Group{ID: "core", Title: "Core:"},
		&cobra.Group{ID: "config", Title: "Configuration:"},
		&cobra.Group{ID: "data", Title: "Data & AI:"},
		&cobra.Group{ID: "infra", Title: "Infrastructure:"},
	)

	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(healthCmd())
	rootCmd.AddCommand(configCmd())

	doctorCmd := doctor.NewCommand()
	doctorCmd.GroupID = "config"
	rootCmd.AddCommand(doctorCmd)

	onboardCmd := onboard.NewCommand()
	onboardCmd.GroupID = "config"
	rootCmd.AddCommand(onboardCmd)

	settingsCmd := settings.NewCommand()
	settingsCmd.GroupID = "config"
	rootCmd.AddCommand(settingsCmd)

	securityCmd := clisecurity.NewSecurityCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	securityCmd.GroupID = "infra"
	rootCmd.AddCommand(securityCmd)

	memoryCmd := climemory.NewMemoryCmd(func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	})
	memoryCmd.GroupID = "data"
	rootCmd.AddCommand(memoryCmd)

	agentCmd := cliagent.NewAgentCmd(func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	})
	agentCmd.GroupID = "data"
	rootCmd.AddCommand(agentCmd)

	graphCmd := cligraph.NewGraphCmd(func() (*config.Config, error) {
		boot, err := bootstrap.Run(bootstrap.Options{})
		if err != nil {
			return nil, err
		}
		defer boot.DBClient.Close()
		return boot.Config, nil
	})
	graphCmd.GroupID = "data"
	rootCmd.AddCommand(graphCmd)

	paymentCmd := clipayment.NewPaymentCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	paymentCmd.GroupID = "infra"
	rootCmd.AddCommand(paymentCmd)

	p2pCmd := clip2p.NewP2PCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	p2pCmd.GroupID = "infra"
	rootCmd.AddCommand(p2pCmd)

	cronCmd := clicron.NewCronCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	cronCmd.GroupID = "infra"
	rootCmd.AddCommand(cronCmd)

	workflowCmd := cliworkflow.NewWorkflowCmd(func() (*bootstrap.Result, error) {
		return bootstrap.Run(bootstrap.Options{})
	})
	workflowCmd.GroupID = "infra"
	rootCmd.AddCommand(workflowCmd)

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
		GroupID: "core",
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
			defer logging.Sync()

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
		GroupID: "core",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("lango %s (built %s)\n", Version, BuildTime)
		},
	}
}

func healthCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:     "health",
		Short:   "Check gateway health (replaces curl in Docker HEALTHCHECK)",
		GroupID: "core",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := "http://localhost:" + strconv.Itoa(port) + "/health"
			client := &http.Client{Timeout: 5 * time.Second}

			resp, err := client.Get(url)
			if err != nil {
				return fmt.Errorf("health check: %w", err)
			}
			defer resp.Body.Close()

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
		GroupID: "config",
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
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new profile with default configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

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

			cfg := config.DefaultConfig()
			if err := boot.ConfigStore.Save(ctx, name, cfg); err != nil {
				return fmt.Errorf("create profile: %w", err)
			}

			fmt.Printf("Profile %q created with default configuration.\n", name)
			return nil
		},
	}
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
				fmt.Scanln(&answer)
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
