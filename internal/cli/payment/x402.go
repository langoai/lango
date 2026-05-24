package payment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newX402Cmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "x402",
		Short:         "Show X402 protocol configuration and auto-pay settings",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			cfg := boot.Config.Payment

			maxAutoPay := cfg.X402.MaxAutoPayAmount
			if maxAutoPay == "" {
				maxAutoPay = "unlimited"
			}

			type x402Output struct {
				PaymentEnabled bool   `json:"payment_enabled"`
				AutoIntercept  bool   `json:"auto_intercept"`
				MaxAutoPayUSDC string `json:"max_auto_pay_usdc"`
				MaxPerTx       string `json:"max_per_tx,omitempty"`
				MaxDaily       string `json:"max_daily,omitempty"`
			}

			out := x402Output{
				PaymentEnabled: cfg.Enabled,
				AutoIntercept:  cfg.X402.AutoIntercept,
				MaxAutoPayUSDC: maxAutoPay,
				MaxPerTx:       cfg.Limits.MaxPerTx,
				MaxDaily:       cfg.Limits.MaxDaily,
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), out)
			}

			autoLabel := "disabled"
			if out.AutoIntercept {
				autoLabel = "enabled"
			}

			fmt.Fprintln(cmd.OutOrStdout(), "X402 Protocol Configuration")
			fmt.Fprintf(cmd.OutOrStdout(), "  Payment Enabled:     %v\n", out.PaymentEnabled)
			fmt.Fprintf(cmd.OutOrStdout(), "  Auto-Intercept:      %s\n", autoLabel)
			fmt.Fprintf(cmd.OutOrStdout(), "  Max Auto-Pay:        %s USDC\n", out.MaxAutoPayUSDC)
			if out.MaxPerTx != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  Max Per Transaction: %s USDC\n", out.MaxPerTx)
			}
			if out.MaxDaily != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  Max Daily Spend:     %s USDC\n", out.MaxDaily)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")

	return cmd
}
