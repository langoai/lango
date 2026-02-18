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

func newInfoCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show wallet and payment system information",
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

			addr, err := deps.service.WalletAddress(ctx)
			if err != nil {
				return fmt.Errorf("get address: %w", err)
			}

			chainID := deps.service.ChainID()
			network := wallet.NetworkName(chainID)

			x402Status := "disabled"
			if deps.config.X402.AutoIntercept {
				x402Status = "enabled (V2 SDK)"
			}
			x402MaxAutoPay := deps.config.X402.MaxAutoPayAmount
			if x402MaxAutoPay == "" {
				x402MaxAutoPay = "unlimited"
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"address":        addr,
					"chainId":        chainID,
					"network":        network,
					"walletProvider": deps.config.WalletProvider,
					"usdcContract":   deps.config.Network.USDCContract,
					"rpcUrl":         deps.config.Network.RPCURL,
					"x402": map[string]interface{}{
						"status":          x402Status,
						"protocol":        "X402 V2 (Coinbase SDK)",
						"maxAutoPayAmount": x402MaxAutoPay,
					},
				})
			}

			fmt.Println("Payment System Info")
			fmt.Printf("  Wallet Address:      %s\n", addr)
			fmt.Printf("  Network:             %s (chain %d)\n", network, chainID)
			fmt.Printf("  Wallet Provider:     %s\n", deps.config.WalletProvider)
			fmt.Printf("  USDC Contract:       %s\n", deps.config.Network.USDCContract)
			fmt.Printf("  RPC URL:             %s\n", deps.config.Network.RPCURL)
			fmt.Printf("  X402 Auto-Intercept: %s\n", x402Status)
			fmt.Printf("  X402 Max Auto-Pay:   %s USDC\n", x402MaxAutoPay)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
