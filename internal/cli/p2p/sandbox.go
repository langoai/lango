package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/sandbox"
)

type namedSandboxExecutor interface {
	sandbox.Executor
	RuntimeName() string
}

type sandboxCleanupRuntime interface {
	IsAvailable(ctx context.Context) bool
	Cleanup(ctx context.Context, id string) error
}

var newContainerSandboxExecutor = func(cfg sandbox.Config, containerCfg config.ContainerSandboxConfig) (namedSandboxExecutor, error) {
	return sandbox.NewContainerExecutor(cfg, containerCfg)
}

var newSubprocessSandboxExecutor = func(cfg sandbox.Config) sandbox.Executor {
	return sandbox.NewSubprocessExecutor(cfg)
}

var newSandboxDockerRuntime = func() (sandboxCleanupRuntime, error) {
	return sandbox.NewDockerRuntime()
}

func newSandboxCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sandbox",
		Short: "Manage P2P tool execution sandbox",
		Long:  "Inspect sandbox status, run smoke tests, and clean up orphaned containers.",
	}

	cmd.AddCommand(newSandboxStatusCmd(bootLoader))
	cmd.AddCommand(newSandboxTestCmd(bootLoader))
	cmd.AddCommand(newSandboxCleanupCmd(bootLoader))

	return cmd
}

func newSandboxStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show sandbox runtime status",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}

			cfg := boot.Config
			if !cfg.P2P.ToolIsolation.Enabled {
				fmt.Fprintln(cmd.OutOrStdout(), "Tool isolation: disabled")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Tool isolation: enabled")
			fmt.Fprintf(cmd.OutOrStdout(), "  Timeout per tool: %v\n", cfg.P2P.ToolIsolation.TimeoutPerTool)
			fmt.Fprintf(cmd.OutOrStdout(), "  Max memory (MB):  %d\n", cfg.P2P.ToolIsolation.MaxMemoryMB)

			if !cfg.P2P.ToolIsolation.Container.Enabled {
				fmt.Fprintln(cmd.OutOrStdout(), "  Container mode:   disabled (subprocess fallback)")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "  Container mode:   enabled")
			fmt.Fprintf(cmd.OutOrStdout(), "  Runtime config:   %s\n", cfg.P2P.ToolIsolation.Container.Runtime)
			fmt.Fprintf(cmd.OutOrStdout(), "  Image:            %s\n", cfg.P2P.ToolIsolation.Container.Image)
			fmt.Fprintf(cmd.OutOrStdout(), "  Network mode:     %s\n", cfg.P2P.ToolIsolation.Container.NetworkMode)

			// Probe actual runtime availability.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			sbxCfg := sandbox.Config{
				Enabled:        true,
				TimeoutPerTool: cfg.P2P.ToolIsolation.TimeoutPerTool,
				MaxMemoryMB:    cfg.P2P.ToolIsolation.MaxMemoryMB,
			}
			exec, err := newContainerSandboxExecutor(sbxCfg, cfg.P2P.ToolIsolation.Container)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  Active runtime:   unavailable (%v)\n", err)
				return nil
			}
			_ = ctx
			fmt.Fprintf(cmd.OutOrStdout(), "  Active runtime:   %s\n", exec.RuntimeName())
			fmt.Fprintf(cmd.OutOrStdout(), "  Pool size:        %d\n", cfg.P2P.ToolIsolation.Container.PoolSize)

			return nil
		},
	}
}

func newSandboxTestCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Run a sandbox smoke test",
		Long:  "Execute a simple echo tool through the sandbox to verify it works.",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}

			cfg := boot.Config
			if !cfg.P2P.ToolIsolation.Enabled {
				return fmt.Errorf("tool isolation is not enabled (set p2p.toolIsolation.enabled = true)")
			}

			sbxCfg := sandbox.Config{
				Enabled:        true,
				TimeoutPerTool: cfg.P2P.ToolIsolation.TimeoutPerTool,
				MaxMemoryMB:    cfg.P2P.ToolIsolation.MaxMemoryMB,
			}

			var exec sandbox.Executor
			if cfg.P2P.ToolIsolation.Container.Enabled {
				containerExec, cErr := newContainerSandboxExecutor(sbxCfg, cfg.P2P.ToolIsolation.Container)
				if cErr != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "Container sandbox unavailable, using subprocess: %v\n", cErr)
					exec = newSubprocessSandboxExecutor(sbxCfg)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Using container runtime: %s\n", containerExec.RuntimeName())
					exec = containerExec
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Using subprocess sandbox")
				exec = newSubprocessSandboxExecutor(sbxCfg)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			params := map[string]interface{}{"msg": "sandbox-smoke-test"}
			result, err := exec.Execute(ctx, "echo", params)
			if err != nil {
				return fmt.Errorf("smoke test: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Smoke test passed: %v\n", result)
			return nil
		},
	}
}

func newSandboxCleanupCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Remove orphaned sandbox containers",
		Long:  "Find and remove Docker containers with label lango.sandbox=true.",
		RunE: func(cmd *cobra.Command, args []string) error {
			dr, err := newSandboxDockerRuntime()
			if err != nil {
				return fmt.Errorf("docker unavailable: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if !dr.IsAvailable(ctx) {
				return fmt.Errorf("docker daemon is not reachable")
			}

			if err := dr.Cleanup(ctx, ""); err != nil {
				return fmt.Errorf("cleanup: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Orphaned sandbox containers cleaned up.")
			return nil
		},
	}
}
