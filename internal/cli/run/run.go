// Package run provides CLI commands for RunLedger (Task OS) management.
package run

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/runledger"
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
	cmd.AddCommand(newJournalCmd(bootLoader))

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

			store := runledger.NewEntStore(boot.DBClient)
			runs, err := store.ListRuns(context.Background(), boot.Config.RunLedger.MaxRunHistory)
			if err != nil {
				return fmt.Errorf("list runs: %w", err)
			}
			if len(runs) == 0 {
				fmt.Println("No runs found.")
				return nil
			}
			for _, run := range runs {
				fmt.Printf("%s\t%s\t%s\t%d/%d\n",
					run.RunID, run.Status, run.Goal, run.CompletedSteps, run.TotalSteps)
			}
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

func newJournalCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "journal <run-id>",
		Short: "View run journal events",
		Args:  cobra.ExactArgs(1),
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

			store := runledger.NewEntStore(boot.DBClient)
			events, err := store.GetJournalEvents(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("get journal events: %w", err)
			}
			for _, event := range events {
				fmt.Printf("%d\t%s\t%s\n", event.Seq, event.Type, string(event.Payload))
			}
			return nil
		},
	}
}
