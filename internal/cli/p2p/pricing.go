package p2p

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/wallet"
)

func newPricingCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		toolName string
		output   string
	)

	cmd := &cobra.Command{
		Use:           "pricing",
		Short:         "Show provider-side P2P quote configuration",
		Long:          "Display the current provider-side P2P quote configuration including default per-query price and tool-specific public quote overrides.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			pricing := boot.Config.P2P.Pricing

			if toolName != "" {
				price, ok := pricing.ToolPrices[toolName]
				if !ok {
					price = pricing.PerQuery
				}
				if output == "json" {
					return printJSON(cmd.OutOrStdout(), map[string]interface{}{
						"tool":     toolName,
						"price":    price,
						"currency": wallet.CurrencyUSDC,
					})
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Tool:     %s\n", toolName)
				fmt.Fprintf(cmd.OutOrStdout(), "Price:    %s %s\n", price, wallet.CurrencyUSDC)
				return nil
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), map[string]interface{}{
					"enabled":    pricing.Enabled,
					"perQuery":   pricing.PerQuery,
					"toolPrices": pricing.ToolPrices,
					"currency":   wallet.CurrencyUSDC,
				})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "P2P Pricing Configuration")
			fmt.Fprintf(cmd.OutOrStdout(), "  Enabled:     %v\n", pricing.Enabled)
			fmt.Fprintf(cmd.OutOrStdout(), "  Per Query:   %s %s\n", pricing.PerQuery, wallet.CurrencyUSDC)
			if len(pricing.ToolPrices) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "  Tool Prices:")
				for tool, price := range pricing.ToolPrices {
					fmt.Fprintf(cmd.OutOrStdout(), "    %-30s %s %s\n", tool, price, wallet.CurrencyUSDC)
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "  Tool Prices: (none)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&toolName, "tool", "", "Filter pricing for a specific tool")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
