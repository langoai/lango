package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/wallet"
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
			defer boot.Close()

			if boot.Storage == nil {
				return fmt.Errorf("payment usage unavailable")
			}
			usage, err := boot.Storage.PaymentUsage(context.Background())
			if err != nil {
				return fmt.Errorf("get daily spent: %w", err)
			}
			maxPerTx := depsFromConfig(boot.Config.Payment.Limits.MaxPerTx)
			maxDaily := depsFromConfig(boot.Config.Payment.Limits.MaxDaily)
			spent := usage.DailySpent
			remaining, err := remainingFromUsage(maxDaily, spent)
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"maxPerTx":       maxPerTx,
					"maxDaily":       maxDaily,
					"dailySpent":     spent,
					"dailyRemaining": remaining,
					"currency":       wallet.CurrencyUSDC,
				})
			}

			fmt.Println("Spending Limits")
			fmt.Printf("  Max Per Transaction:  %s %s\n", maxPerTx, wallet.CurrencyUSDC)
			fmt.Printf("  Max Daily:            %s %s\n", maxDaily, wallet.CurrencyUSDC)
			fmt.Printf("  Spent Today:          %s %s\n", spent, wallet.CurrencyUSDC)
			fmt.Printf("  Remaining Today:      %s %s\n", remaining, wallet.CurrencyUSDC)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func depsFromConfig(v string) string {
	if v == "" {
		return "0"
	}
	return v
}

func remainingFromUsage(maxDaily, spent string) (string, error) {
	maxAmt, err := wallet.ParseUSDC(maxDaily)
	if err != nil {
		return "", fmt.Errorf("parse max daily: %w", err)
	}
	spentAmt, err := wallet.ParseUSDC(spent)
	if err != nil {
		return "", fmt.Errorf("parse spent amount: %w", err)
	}
	remaining := new(big.Int).Sub(maxAmt, spentAmt)
	if remaining.Sign() < 0 {
		remaining = big.NewInt(0)
	}
	return wallet.FormatUSDC(remaining), nil
}
