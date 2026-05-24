// Package run provides CLI commands for RunLedger (Task OS) management.
package run

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/runledger"
)

var runLedgerStoreFromBoot = func(boot *bootstrap.Result) runledger.RunLedgerStore {
	if boot != nil && boot.Storage != nil {
		return boot.Storage.RunLedger()
	}
	return nil
}

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
	var (
		output string
		limit  int
	)

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List recent runs",
		Long:          `List recent RunLedger runs from the persistent snapshot store.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			if !boot.Config.RunLedger.Enabled {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "RunLedger is disabled. Enable with: lango config set runLedger.enabled true")
				return err
			}

			store := runLedgerStoreFromBoot(boot)
			if store == nil {
				return fmt.Errorf("runledger store unavailable")
			}
			effectiveLimit := limit
			if effectiveLimit <= 0 {
				effectiveLimit = boot.Config.RunLedger.MaxRunHistory
			}
			runs, err := store.ListRuns(context.Background(), effectiveLimit)
			if err != nil {
				return fmt.Errorf("list runs: %w", err)
			}
			if output == "json" {
				return printJSON(cmd.OutOrStdout(), runs)
			}
			if len(runs) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No runs found.")
				return err
			}
			for _, run := range runs {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%d/%d\n",
					run.RunID, run.Status, run.Goal, run.CompletedSteps, run.TotalSteps)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of runs to list (0 uses config default)")
	return cmd
}

func newStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show RunLedger configuration",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			cfg := boot.Config.RunLedger
			if output == "json" {
				return printJSON(cmd.OutOrStdout(), cfg)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "RunLedger Configuration:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Enabled:            %v\n", cfg.Enabled)
			fmt.Fprintf(cmd.OutOrStdout(), "  Shadow Mode:        %v\n", cfg.Shadow)
			fmt.Fprintf(cmd.OutOrStdout(), "  Write-Through:      %v\n", cfg.WriteThrough)
			fmt.Fprintf(cmd.OutOrStdout(), "  Authoritative Read: %v\n", cfg.AuthoritativeRead)
			fmt.Fprintf(cmd.OutOrStdout(), "  Workspace Isolation:%v\n", cfg.WorkspaceIsolation)
			fmt.Fprintf(cmd.OutOrStdout(), "  Stale TTL:          %v\n", cfg.StaleTTL)
			fmt.Fprintf(cmd.OutOrStdout(), "  Max Run History:    %d\n", cfg.MaxRunHistory)
			fmt.Fprintf(cmd.OutOrStdout(), "  Validator Timeout:  %v\n", cfg.ValidatorTimeout)
			fmt.Fprintf(cmd.OutOrStdout(), "  Planner Retries:    %d\n", cfg.PlannerMaxRetries)
			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}

func newJournalCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		output string
		limit  int
	)

	cmd := &cobra.Command{
		Use:           "journal <run-id>",
		Short:         "View run journal events",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			if !boot.Config.RunLedger.Enabled {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "RunLedger is disabled. Enable with: lango config set runLedger.enabled true")
				return err
			}

			store := runLedgerStoreFromBoot(boot)
			if store == nil {
				return fmt.Errorf("runledger store unavailable")
			}
			events, err := store.GetJournalEvents(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("get journal events: %w", err)
			}
			if limit > 0 && len(events) > limit {
				events = events[:limit]
			}
			if output == "json" {
				return printJSON(cmd.OutOrStdout(), events)
			}
			for _, event := range events {
				fmt.Fprintf(cmd.OutOrStdout(), "%d\t%s\t%s\n", event.Seq, event.Type, string(event.Payload))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of journal events to show (0 means all)")
	return cmd
}
