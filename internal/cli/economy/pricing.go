package economy

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

func newPricingCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pricing",
		Short: "Show dynamic pricing policy configuration",
		Long:  "Display the local dynamic pricing policy configuration that layers above the P2P market quote surface.",
	}

	cmd.AddCommand(newPricingStatusCmd(cfgLoader))
	return cmd
}

func newPricingStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show dynamic pricing policy configuration",
		Long:  "Display the current dynamic pricing policy configuration for the local economy engine.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return err
			}

			if !cfg.Economy.Enabled || !cfg.Economy.Pricing.Enabled {
				fmt.Fprintln(cmd.OutOrStdout(), "Dynamic pricing is disabled.")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Pricing Configuration:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Trust Discount:  %.0f%%\n", cfg.Economy.Pricing.TrustDiscount*100)
			fmt.Fprintf(cmd.OutOrStdout(), "  Volume Discount: %.0f%%\n", cfg.Economy.Pricing.VolumeDiscount*100)
			fmt.Fprintf(cmd.OutOrStdout(), "  Min Price:       %s USDC\n", cfg.Economy.Pricing.MinPrice)
			return nil
		},
	}
}
