package payment

import (
	"context"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/wallet"
)

func newLimitsCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "limits",
		Short:         "Show spending limits and daily usage",
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

			usage, err := paymentUsageLoader(context.Background(), boot)
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

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), map[string]interface{}{
					"maxPerTx":       maxPerTx,
					"maxDaily":       maxDaily,
					"dailySpent":     spent,
					"dailyRemaining": remaining,
					"currency":       wallet.CurrencyUSDC,
				})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Spending Limits")
			fmt.Fprintf(cmd.OutOrStdout(), "  Max Per Transaction:  %s %s\n", maxPerTx, wallet.CurrencyUSDC)
			fmt.Fprintf(cmd.OutOrStdout(), "  Max Daily:            %s %s\n", maxDaily, wallet.CurrencyUSDC)
			fmt.Fprintf(cmd.OutOrStdout(), "  Spent Today:          %s %s\n", spent, wallet.CurrencyUSDC)
			fmt.Fprintf(cmd.OutOrStdout(), "  Remaining Today:      %s %s\n", remaining, wallet.CurrencyUSDC)

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
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
