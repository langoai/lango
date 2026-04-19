package economy

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

func newNegotiateCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "negotiate",
		Short: "Show negotiation engine configuration",
		Long:  "Display the local negotiation engine and configuration layered above the P2P market path.",
	}

	cmd.AddCommand(newNegotiateStatusCmd(cfgLoader))
	return cmd
}

func newNegotiateStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show negotiation engine configuration",
		Long:  "Display the current negotiation engine configuration for the local economy policy layer.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return err
			}

			if !cfg.Economy.Enabled || !cfg.Economy.Negotiate.Enabled {
				fmt.Println("Negotiation is disabled.")
				return nil
			}

			fmt.Println("Negotiation Configuration:")
			fmt.Printf("  Max Rounds:     %d\n", cfg.Economy.Negotiate.MaxRounds)
			fmt.Printf("  Timeout:        %s\n", cfg.Economy.Negotiate.Timeout)
			fmt.Printf("  Auto Negotiate: %v\n", cfg.Economy.Negotiate.AutoNegotiate)
			fmt.Printf("  Max Discount:   %.0f%%\n", cfg.Economy.Negotiate.MaxDiscount*100)
			return nil
		},
	}
}
