package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/langowarny/lango/internal/app"
	"github.com/langowarny/lango/internal/cli/auth"
	"github.com/langowarny/lango/internal/cli/doctor"
	climemory "github.com/langowarny/lango/internal/cli/memory"
	"github.com/langowarny/lango/internal/cli/onboard"
	clisecurity "github.com/langowarny/lango/internal/cli/security"
	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/logging"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	cfgFile   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "lango",
		Short: "Lango - Fast AI Agent in Go",
		Long:  `Lango is a high-performance AI agent built with Go, supporting multiple channels and tools.`,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: lango.json)")

	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(doctor.NewCommand())
	rootCmd.AddCommand(onboard.NewCommand())
	rootCmd.AddCommand(auth.NewCommand())
	rootCmd.AddCommand(clisecurity.NewSecurityCmd(func() (*config.Config, error) {
		return config.Load(cfgFile)
	}))
	rootCmd.AddCommand(climemory.NewMemoryCmd(func() (*config.Config, error) {
		return config.Load(cfgFile)
	}))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the gateway server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Initialize logging
			if err := logging.Init(logging.LogConfig{
				Level:      cfg.Logging.Level,
				Format:     cfg.Logging.Format,
				OutputPath: cfg.Logging.OutputPath,
			}); err != nil {
				return fmt.Errorf("failed to init logging: %w", err)
			}
			defer logging.Sync()

			log := logging.Sugar()
			log.Infow("starting lango", "version", Version)

			// Create application
			application, err := app.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to create application: %w", err)
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
					log.Errorw("shutdown error", "error", err)
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
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("lango %s (built %s)\n", Version, BuildTime)
		},
	}
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("configuration invalid: %w", err)
			}

			if err := config.Validate(cfg); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			fmt.Println("Configuration is valid")
			return nil
		},
	})

	return cmd
}
