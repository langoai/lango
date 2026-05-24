package payment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/wallet"
)

func newBalanceCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "balance",
		Short:         "Show USDC wallet balance",
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

			deps, err := initPaymentDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			ctx := context.Background()

			balance, err := deps.service.Balance(ctx)
			if err != nil {
				return fmt.Errorf("get balance: %w", err)
			}

			addr, err := deps.service.WalletAddress(ctx)
			if err != nil {
				return fmt.Errorf("get address: %w", err)
			}

			chainID := deps.service.ChainID()
			network := wallet.NetworkName(chainID)

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), map[string]interface{}{
					"balance":  balance,
					"currency": wallet.CurrencyUSDC,
					"address":  addr,
					"chainId":  chainID,
					"network":  network,
				})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Wallet Balance")
			fmt.Fprintf(cmd.OutOrStdout(), "  Balance:   %s %s\n", balance, wallet.CurrencyUSDC)
			fmt.Fprintf(cmd.OutOrStdout(), "  Address:   %s\n", addr)
			fmt.Fprintf(cmd.OutOrStdout(), "  Network:   %s (chain %d)\n", network, chainID)

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
