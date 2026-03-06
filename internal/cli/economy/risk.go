package economy

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

func newRiskCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "risk",
		Short: "Manage risk assessment",
	}

	cmd.AddCommand(newRiskStatusCmd(cfgLoader))
	return cmd
}

func newRiskStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show risk assessment configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return err
			}

			if !cfg.Economy.Enabled {
				fmt.Println("Economy layer is disabled.")
				return nil
			}

			fmt.Println("Risk Configuration:")
			fmt.Printf("  Escrow Threshold: %s USDC\n", cfg.Economy.Risk.EscrowThreshold)
			fmt.Printf("  High Trust Score: %.2f\n", cfg.Economy.Risk.HighTrustScore)
			fmt.Printf("  Med Trust Score:  %.2f\n", cfg.Economy.Risk.MediumTrustScore)
			return nil
		},
	}
}
