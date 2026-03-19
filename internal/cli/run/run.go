// Package run provides CLI commands for RunLedger (Task OS) management.
package run

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

// NewRunCmd creates the run command with lazy bootstrap loading.
func NewRunCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Manage RunLedger (Task OS) runs",
		Long:  "List, inspect, and manage durable execution runs powered by the RunLedger engine.",
	}

	cmd.AddCommand(newListCmd(bootLoader))
	cmd.AddCommand(newStatusCmd(bootLoader))
	cmd.AddCommand(newJournalCmd())

	return cmd
}

func newListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List recent runs",
		Long: `List recent RunLedger runs.

Phase 1 uses an in-memory store, so runs are only available during the
current server session. Persistent storage is introduced in Phase 2.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			if !boot.Config.RunLedger.Enabled {
				fmt.Println("RunLedger is disabled. Enable with: lango config set runLedger.enabled true")
				return nil
			}

			fmt.Println("RunLedger is enabled (Phase 1: in-memory store).")
			fmt.Println("Runs are available only during the current server session.")
			fmt.Println("Use 'lango serve' to start the server and interact with runs via agent tools.")
			return nil
		},
	}
}

func newStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show RunLedger configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			cfg := boot.Config.RunLedger
			fmt.Printf("RunLedger Configuration:\n")
			fmt.Printf("  Enabled:            %v\n", cfg.Enabled)
			fmt.Printf("  Shadow Mode:        %v\n", cfg.Shadow)
			fmt.Printf("  Write-Through:      %v\n", cfg.WriteThrough)
			fmt.Printf("  Authoritative Read: %v\n", cfg.AuthoritativeRead)
			fmt.Printf("  Stale TTL:          %v\n", cfg.StaleTTL)
			fmt.Printf("  Max Run History:    %d\n", cfg.MaxRunHistory)
			fmt.Printf("  Validator Timeout:  %v\n", cfg.ValidatorTimeout)
			fmt.Printf("  Planner Retries:    %d\n", cfg.PlannerMaxRetries)
			return nil
		},
	}
}

func newJournalCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "journal",
		Short: "View run journal events",
		Long: `View the journal event log for a specific run.

This command requires persistent storage which is introduced in Phase 2.
In Phase 1 (in-memory store), journal data is only available during the
current server session via agent tools.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Journal viewing requires persistent store (Phase 2).")
			fmt.Println("In Phase 1, use run_read tool via the agent to inspect run state.")
			return nil
		},
	}
}
