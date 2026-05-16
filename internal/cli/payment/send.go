package payment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/prompt"
	pmtypes "github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/wallet"
)

func newSendCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		to      string
		amount  string
		purpose string
		force   bool
		output  string
	)

	cmd := &cobra.Command{
		Use:           "send",
		Short:         "Send a USDC payment",
		Long:          "Send USDC to a recipient address on the configured blockchain network.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			if to == "" || amount == "" || purpose == "" {
				return fmt.Errorf("--to, --amount, and --purpose are required")
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			deps, err := paymentDepsLoader(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			chainID := deps.service.ChainID()
			network := wallet.NetworkName(chainID)

			// Confirmation prompt unless --force.
			if !force {
				if err := prompt.RequireTTYInput(cmd.InOrStdin(), "use --force for non-interactive mode"); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Send %s USDC to %s on %s?\n", amount, to, network); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Purpose: %s\n", purpose); err != nil {
					return err
				}
				ok, err := prompt.ConfirmDenyOnEOFIO(cmd.InOrStdin(), cmd.OutOrStdout(), "Confirm")
				if err != nil {
					return err
				}
				if !ok {
					_, err := fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return err
				}
			}

			ctx := context.Background()
			receipt, err := paymentSendExecutor(ctx, deps, pmtypes.PaymentRequest{
				To:      to,
				Amount:  amount,
				Purpose: purpose,
			})
			if err != nil {
				return fmt.Errorf("send payment: %w", err)
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), map[string]interface{}{
					"status":  receipt.Status,
					"txHash":  receipt.TxHash,
					"amount":  receipt.Amount,
					"from":    receipt.From,
					"to":      receipt.To,
					"chainId": receipt.ChainID,
					"network": network,
				})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Payment Submitted")
			fmt.Fprintf(cmd.OutOrStdout(), "  Status:    %s\n", receipt.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "  Tx Hash:   %s\n", receipt.TxHash)
			fmt.Fprintf(cmd.OutOrStdout(), "  Amount:    %s USDC\n", receipt.Amount)
			fmt.Fprintf(cmd.OutOrStdout(), "  From:      %s\n", receipt.From)
			fmt.Fprintf(cmd.OutOrStdout(), "  To:        %s\n", receipt.To)
			fmt.Fprintf(cmd.OutOrStdout(), "  Network:   %s (chain %d)\n", network, receipt.ChainID)

			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Recipient wallet address (0x...)")
	cmd.Flags().StringVar(&amount, "amount", "", "Amount in USDC (e.g. \"1.50\")")
	cmd.Flags().StringVar(&purpose, "purpose", "", "Human-readable purpose of the payment")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")

	return cmd
}
