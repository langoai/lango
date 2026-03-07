package economy

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

func newEscrowCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "escrow",
		Short: "Manage escrow agreements",
	}

	cmd.AddCommand(newEscrowStatusCmd(cfgLoader))
	return cmd
}

func newEscrowStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show escrow configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return err
			}

			if !cfg.Economy.Enabled || !cfg.Economy.Escrow.Enabled {
				fmt.Println("Escrow is disabled.")
				return nil
			}

			fmt.Println("Escrow Configuration:")
			fmt.Printf("  Default Timeout: %s\n", cfg.Economy.Escrow.DefaultTimeout)
			fmt.Printf("  Max Milestones:  %d\n", cfg.Economy.Escrow.MaxMilestones)
			fmt.Printf("  Auto Release:    %v\n", cfg.Economy.Escrow.AutoRelease)
			fmt.Printf("  Dispute Window:  %s\n", cfg.Economy.Escrow.DisputeWindow)
			return nil
		},
	}
}
