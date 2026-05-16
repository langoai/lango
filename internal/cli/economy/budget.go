package economy

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

func newBudgetCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Manage task budgets",
	}

	cmd.AddCommand(newBudgetStatusCmd(cfgLoader))
	return cmd
}

func newBudgetStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var taskID string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show budget configuration and status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return err
			}

			if !cfg.Economy.Enabled {
				fmt.Fprintln(cmd.OutOrStdout(), "Economy layer is disabled. Enable with economy.enabled=true")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Budget Configuration:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Default Max:      %s USDC\n", cfg.Economy.Budget.DefaultMax)
			fmt.Fprintf(cmd.OutOrStdout(), "  Alert Thresholds: %v\n", cfg.Economy.Budget.AlertThresholds)
			if cfg.Economy.Budget.HardLimit == nil || *cfg.Economy.Budget.HardLimit {
				fmt.Fprintln(cmd.OutOrStdout(), "  Hard Limit:       enabled")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "  Hard Limit:       disabled")
			}

			if taskID != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "\nTask %q budget: use 'lango serve' and economy_budget_status tool for live data\n", taskID)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task-id", "", "Task ID to check")
	return cmd
}
