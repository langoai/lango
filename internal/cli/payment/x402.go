package payment

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newX402Cmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "x402",
		Short: "Show X402 protocol configuration and auto-pay settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

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

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			autoLabel := "disabled"
			if out.AutoIntercept {
				autoLabel = "enabled"
			}

			fmt.Println("X402 Protocol Configuration")
			fmt.Printf("  Payment Enabled:     %v\n", out.PaymentEnabled)
			fmt.Printf("  Auto-Intercept:      %s\n", autoLabel)
			fmt.Printf("  Max Auto-Pay:        %s USDC\n", out.MaxAutoPayUSDC)
			if out.MaxPerTx != "" {
				fmt.Printf("  Max Per Transaction: %s USDC\n", out.MaxPerTx)
			}
			if out.MaxDaily != "" {
				fmt.Printf("  Max Daily Spend:     %s USDC\n", out.MaxDaily)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
