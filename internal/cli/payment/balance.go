package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/wallet"
)

func newBalanceCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Show USDC wallet balance",
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

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"balance":  balance,
					"currency": "USDC",
					"address":  addr,
					"chainId":  chainID,
					"network":  network,
				})
			}

			fmt.Println("Wallet Balance")
			fmt.Printf("  Balance:   %s USDC\n", balance)
			fmt.Printf("  Address:   %s\n", addr)
			fmt.Printf("  Network:   %s (chain %d)\n", network, chainID)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
