package p2p

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newPricingCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		toolName   string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "pricing",
		Short: "Show P2P tool pricing configuration",
		Long:  "Display the current P2P pricing configuration including default per-query price and tool-specific price overrides.",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			pricing := boot.Config.P2P.Pricing

			if toolName != "" {
				price, ok := pricing.ToolPrices[toolName]
				if !ok {
					price = pricing.PerQuery
				}
				if jsonOutput {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(map[string]interface{}{
						"tool":     toolName,
						"price":    price,
						"currency": "USDC",
					})
				}
				fmt.Printf("Tool:     %s\n", toolName)
				fmt.Printf("Price:    %s USDC\n", price)
				return nil
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"enabled":    pricing.Enabled,
					"perQuery":   pricing.PerQuery,
					"toolPrices": pricing.ToolPrices,
					"currency":   "USDC",
				})
			}

			fmt.Println("P2P Pricing Configuration")
			fmt.Printf("  Enabled:     %v\n", pricing.Enabled)
			fmt.Printf("  Per Query:   %s USDC\n", pricing.PerQuery)
			if len(pricing.ToolPrices) > 0 {
				fmt.Println("  Tool Prices:")
				for tool, price := range pricing.ToolPrices {
					fmt.Printf("    %-30s %s USDC\n", tool, price)
				}
			} else {
				fmt.Println("  Tool Prices: (none)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&toolName, "tool", "", "Filter pricing for a specific tool")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
