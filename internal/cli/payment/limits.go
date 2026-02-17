package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langowarny/lango/internal/bootstrap"
	"github.com/langowarny/lango/internal/wallet"
)

func newLimitsCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "limits",
		Short: "Show spending limits and daily usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			deps, err := initPaymentDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			ctx := context.Background()

			spent, err := deps.limiter.DailySpent(ctx)
			if err != nil {
				return fmt.Errorf("get daily spent: %w", err)
			}

			remaining, err := deps.limiter.DailyRemaining(ctx)
			if err != nil {
				return fmt.Errorf("get daily remaining: %w", err)
			}

			maxPerTx := deps.limiter.MaxPerTx()
			maxDaily := deps.limiter.MaxDaily()

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"maxPerTx":       wallet.FormatUSDC(maxPerTx),
					"maxDaily":       wallet.FormatUSDC(maxDaily),
					"dailySpent":     wallet.FormatUSDC(spent),
					"dailyRemaining": wallet.FormatUSDC(remaining),
					"currency":       "USDC",
				})
			}

			fmt.Println("Spending Limits")
			fmt.Printf("  Max Per Transaction:  %s USDC\n", wallet.FormatUSDC(maxPerTx))
			fmt.Printf("  Max Daily:            %s USDC\n", wallet.FormatUSDC(maxDaily))
			fmt.Printf("  Spent Today:          %s USDC\n", wallet.FormatUSDC(spent))
			fmt.Printf("  Remaining Today:      %s USDC\n", wallet.FormatUSDC(remaining))

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
