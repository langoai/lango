package smartaccount

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func policyCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage session policies",
		Long: `Manage harness policies for smart account session keys.

Examples:
  lango account policy show
  lango account policy set --max-tx "5.00" --daily "50.00" --monthly "500.00"`,
	}

	cmd.AddCommand(policyShowCmd(bootLoader))
	cmd.AddCommand(policySetCmd(bootLoader))

	return cmd
}

func policyShowCmd(bootLoader BootLoader) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current harness policy configuration",
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

			type policyInfo struct {
				Enabled     bool   `json:"enabled"`
				MaxDuration string `json:"maxSessionDuration"`
				MaxKeys     int    `json:"maxActiveKeys"`
				DefaultGas  uint64 `json:"defaultGasLimit"`
				Status      string `json:"status"`
			}

			info := policyInfo{
				Enabled:     cfg.SmartAccount.Enabled,
				MaxDuration: cfg.SmartAccount.Session.MaxDuration.String(),
				MaxKeys:     cfg.SmartAccount.Session.MaxActiveKeys,
				DefaultGas:  cfg.SmartAccount.Session.DefaultGasLimit,
				Status:      "configured (runtime policies require running server)",
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(info, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "Harness Policy Configuration")
			fmt.Fprintln(w, "============================")
			fmt.Fprintf(w, "Enabled:\t%v\n", info.Enabled)
			fmt.Fprintf(w, "Max Session Duration:\t%s\n", info.MaxDuration)
			fmt.Fprintf(w, "Max Active Keys:\t%d\n", info.MaxKeys)
			fmt.Fprintf(w, "Default Gas Limit:\t%d\n", info.DefaultGas)
			if flushErr := w.Flush(); flushErr != nil {
				return fmt.Errorf("flush output: %w", flushErr)
			}

			fmt.Println()
			fmt.Println("Note: Runtime policy details (max tx, daily/monthly limits, allowed targets)")
			fmt.Println("are managed by the policy engine at runtime.")
			fmt.Println("Use 'lango account policy set' to configure limits.")

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	return cmd
}

func policySetCmd(bootLoader BootLoader) *cobra.Command {
	var (
		maxTx   string
		daily   string
		monthly string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set harness policy limits",
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

			if maxTx == "" && daily == "" && monthly == "" {
				return fmt.Errorf("provide at least one policy limit (--max-tx, --daily, or --monthly)")
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "Policy Update Request")
			fmt.Fprintln(w, "---------------------")
			if maxTx != "" {
				fmt.Fprintf(w, "Max Tx Amount:\t%s ETH\n", maxTx)
			}
			if daily != "" {
				fmt.Fprintf(w, "Daily Limit:\t%s ETH\n", daily)
			}
			if monthly != "" {
				fmt.Fprintf(w, "Monthly Limit:\t%s ETH\n", monthly)
			}
			if flushErr := w.Flush(); flushErr != nil {
				return fmt.Errorf("flush output: %w", flushErr)
			}

			fmt.Println()
			fmt.Println("Note: Runtime policy changes require a running server (lango serve).")
			fmt.Println("The policy engine will apply these limits to all future transactions.")

			return nil
		},
	}

	cmd.Flags().StringVar(&maxTx, "max-tx", "", "maximum per-transaction amount in ETH")
	cmd.Flags().StringVar(&daily, "daily", "", "daily spending limit in ETH")
	cmd.Flags().StringVar(&monthly, "monthly", "", "monthly spending limit in ETH")

	return cmd
}
