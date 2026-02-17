package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langowarny/lango/internal/bootstrap"
	"github.com/langowarny/lango/internal/cli/prompt"
	pmtypes "github.com/langowarny/lango/internal/payment"
	"github.com/langowarny/lango/internal/wallet"
)

func newSendCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		to         string
		amount     string
		purpose    string
		force      bool
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a USDC payment",
		Long:  "Send USDC to a recipient address on the configured blockchain network.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if to == "" || amount == "" || purpose == "" {
				return fmt.Errorf("--to, --amount, and --purpose are required")
			}

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

			chainID := deps.service.ChainID()
			network := wallet.NetworkName(chainID)

			// Confirmation prompt unless --force.
			if !force {
				if !prompt.IsInteractive() {
					return fmt.Errorf("use --force for non-interactive mode")
				}
				fmt.Printf("Send %s USDC to %s on %s?\n", amount, to, network)
				fmt.Printf("Purpose: %s\n", purpose)
				fmt.Print("Confirm [y/N]: ")
				var answer string
				fmt.Scanln(&answer)
				if answer != "y" && answer != "Y" && answer != "yes" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			ctx := context.Background()
			receipt, err := deps.service.Send(ctx, pmtypes.PaymentRequest{
				To:      to,
				Amount:  amount,
				Purpose: purpose,
			})
			if err != nil {
				return fmt.Errorf("send payment: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"status":  receipt.Status,
					"txHash":  receipt.TxHash,
					"amount":  receipt.Amount,
					"from":    receipt.From,
					"to":      receipt.To,
					"chainId": receipt.ChainID,
					"network": network,
				})
			}

			fmt.Println("Payment Submitted")
			fmt.Printf("  Status:    %s\n", receipt.Status)
			fmt.Printf("  Tx Hash:   %s\n", receipt.TxHash)
			fmt.Printf("  Amount:    %s USDC\n", receipt.Amount)
			fmt.Printf("  From:      %s\n", receipt.From)
			fmt.Printf("  To:        %s\n", receipt.To)
			fmt.Printf("  Network:   %s (chain %d)\n", network, receipt.ChainID)

			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Recipient wallet address (0x...)")
	cmd.Flags().StringVar(&amount, "amount", "", "Amount in USDC (e.g. \"1.50\")")
	cmd.Flags().StringVar(&purpose, "purpose", "", "Human-readable purpose of the payment")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
