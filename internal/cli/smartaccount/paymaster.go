package smartaccount

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func paymasterCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paymaster",
		Short: "Manage ERC-4337 paymaster for gasless USDC transactions",
	}

	cmd.AddCommand(paymasterStatusCmd(bootLoader))
	cmd.AddCommand(paymasterApproveCmd(bootLoader))

	return cmd
}

func paymasterStatusCmd(bootLoader BootLoader) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show paymaster configuration and approval status",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			cfg := boot.Config
			if !cfg.SmartAccount.Enabled {
				return fmt.Errorf("smart account not enabled in config")
			}

			pmCfg := cfg.SmartAccount.Paymaster

			type statusInfo struct {
				Enabled          bool   `json:"enabled"`
				Provider         string `json:"provider"`
				RPCURL           string `json:"rpcURL"`
				TokenAddress     string `json:"tokenAddress"`
				PaymasterAddress string `json:"paymasterAddress"`
				PolicyID         string `json:"policyId,omitempty"`
			}

			info := statusInfo{
				Enabled:          pmCfg.Enabled,
				Provider:         pmCfg.Provider,
				RPCURL:           pmCfg.RPCURL,
				TokenAddress:     pmCfg.TokenAddress,
				PaymasterAddress: pmCfg.PaymasterAddress,
				PolicyID:         pmCfg.PolicyID,
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(info, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println("Paymaster Status")
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "  Enabled:\t%v\n", info.Enabled)
			fmt.Fprintf(w, "  Provider:\t%s\n", info.Provider)
			fmt.Fprintf(w, "  RPC URL:\t%s\n", info.RPCURL)
			fmt.Fprintf(w, "  Token:\t%s\n", info.TokenAddress)
			fmt.Fprintf(w, "  Paymaster:\t%s\n", info.PaymasterAddress)
			if info.PolicyID != "" {
				fmt.Fprintf(w, "  Policy ID:\t%s\n", info.PolicyID)
			}
			if flushErr := w.Flush(); flushErr != nil {
				return fmt.Errorf("flush output: %w", flushErr)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	return cmd
}

func paymasterApproveCmd(bootLoader BootLoader) *cobra.Command {
	var (
		output string
		amount string
	)

	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve USDC spending for the paymaster",
		Long: `Approve the paymaster to spend USDC from your smart account.
This is required before the paymaster can sponsor gas in USDC.

Examples:
  lango account paymaster approve --amount 1000.00
  lango account paymaster approve --amount max`,
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			cfg := boot.Config
			if !cfg.SmartAccount.Enabled {
				return fmt.Errorf("smart account not enabled in config")
			}
			if !cfg.SmartAccount.Paymaster.Enabled {
				return fmt.Errorf("paymaster not enabled in config")
			}

			type approveInfo struct {
				Token     string `json:"token"`
				Paymaster string `json:"paymaster"`
				Amount    string `json:"amount"`
				Note      string `json:"note"`
			}

			info := approveInfo{
				Token:     cfg.SmartAccount.Paymaster.TokenAddress,
				Paymaster: cfg.SmartAccount.Paymaster.PaymasterAddress,
				Amount:    amount,
				Note:      "Use the 'paymaster_approve' agent tool for actual on-chain approval.",
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(info, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println("Paymaster USDC Approval")
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "  Token:\t%s\n", info.Token)
			fmt.Fprintf(w, "  Paymaster:\t%s\n", info.Paymaster)
			fmt.Fprintf(w, "  Amount:\t%s USDC\n", info.Amount)
			if flushErr := w.Flush(); flushErr != nil {
				return fmt.Errorf("flush output: %w", flushErr)
			}

			fmt.Println()
			fmt.Println("Note: Full approval requires a running server (lango serve).")
			fmt.Println("Use the 'paymaster_approve' agent tool for actual on-chain approval.")

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	cmd.Flags().StringVar(&amount, "amount", "1000.00", "USDC amount to approve (or 'max' for unlimited)")
	return cmd
}
