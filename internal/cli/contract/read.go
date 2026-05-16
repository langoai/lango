package contract

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	contractpkg "github.com/langoai/lango/internal/contract"
)

func newReadCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		address string
		abiFile string
		method  string
		argsStr string
		chainID int64
		output  string
	)

	cmd := &cobra.Command{
		Use:           "read",
		Short:         "Read data from a smart contract (view/pure call)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			if !cfg.Payment.Enabled {
				return fmt.Errorf("payment system is not enabled (set payment.enabled = true)")
			}

			abiJSON, err := os.ReadFile(abiFile)
			if err != nil {
				return fmt.Errorf("read ABI file %q: %w", abiFile, err)
			}

			if chainID == 0 {
				chainID = cfg.Payment.Network.ChainID
			}

			var callArgs []interface{}
			if argsStr != "" {
				for _, a := range strings.Split(argsStr, ",") {
					callArgs = append(callArgs, strings.TrimSpace(a))
				}
			}

			// Create a minimal caller for CLI use.
			// Full RPC is not established here — this is a config-only subcommand.
			// The actual contract read requires a running server or full bootstrap.
			cache := contractpkg.NewABICache()
			parsed, err := cache.GetOrParse(chainID, common.HexToAddress(address), string(abiJSON))
			if err != nil {
				return fmt.Errorf("parse ABI: %w", err)
			}

			// Validate method exists.
			if _, ok := parsed.Methods[method]; !ok {
				return fmt.Errorf("method %q not found in ABI", method)
			}

			fmt.Fprintln(cmd.ErrOrStderr(), "Note: contract read requires a running RPC connection.")
			fmt.Fprintln(cmd.ErrOrStderr(), "Use 'lango serve' and the contract_read agent tool for live queries.")
			fmt.Fprintln(cmd.ErrOrStderr())

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), map[string]interface{}{
					"address": address,
					"method":  method,
					"args":    callArgs,
					"chainId": chainID,
					"status":  "validated",
				})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Contract Read (validated)")
			fmt.Fprintf(cmd.OutOrStdout(), "  Address:  %s\n", address)
			fmt.Fprintf(cmd.OutOrStdout(), "  Method:   %s\n", method)
			if len(callArgs) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  Args:     %v\n", callArgs)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Chain ID: %d\n", chainID)

			return nil
		},
	}

	cmd.Flags().StringVar(&address, "address", "", "Contract address (0x...)")
	cmd.Flags().StringVar(&abiFile, "abi", "", "Path to ABI JSON file")
	cmd.Flags().StringVar(&method, "method", "", "Method name to call")
	cmd.Flags().StringVar(&argsStr, "args", "", "Comma-separated method arguments")
	cmd.Flags().Int64Var(&chainID, "chain-id", 0, "Chain ID (default: from config)")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")

	_ = cmd.MarkFlagRequired("address")
	_ = cmd.MarkFlagRequired("abi")
	_ = cmd.MarkFlagRequired("method")

	return cmd
}
