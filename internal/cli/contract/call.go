package contract

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	contractpkg "github.com/langoai/lango/internal/contract"
)

func newCallCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		address string
		abiFile string
		method  string
		argsStr string
		value   string
		chainID int64
		asJSON  bool
	)

	cmd := &cobra.Command{
		Use:   "call",
		Short: "Send a state-changing transaction to a smart contract",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			cache := contractpkg.NewABICache()
			parsed, err := cache.GetOrParse(chainID, common.HexToAddress(address), string(abiJSON))
			if err != nil {
				return fmt.Errorf("parse ABI: %w", err)
			}

			if _, ok := parsed.Methods[method]; !ok {
				return fmt.Errorf("method %q not found in ABI", method)
			}

			fmt.Fprintf(os.Stderr, "Note: contract call requires a running RPC connection and wallet.\n")
			fmt.Fprintf(os.Stderr, "Use 'lango serve' and the contract_call agent tool for live transactions.\n\n")

			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"address": address,
					"method":  method,
					"args":    callArgs,
					"value":   value,
					"chainId": chainID,
					"status":  "validated",
				})
			}

			fmt.Printf("Contract Call (validated)\n")
			fmt.Printf("  Address:  %s\n", address)
			fmt.Printf("  Method:   %s\n", method)
			if len(callArgs) > 0 {
				fmt.Printf("  Args:     %v\n", callArgs)
			}
			if value != "" {
				fmt.Printf("  Value:    %s ETH\n", value)
			}
			fmt.Printf("  Chain ID: %d\n", chainID)

			return nil
		},
	}

	cmd.Flags().StringVar(&address, "address", "", "Contract address (0x...)")
	cmd.Flags().StringVar(&abiFile, "abi", "", "Path to ABI JSON file")
	cmd.Flags().StringVar(&method, "method", "", "Method name to call")
	cmd.Flags().StringVar(&argsStr, "args", "", "Comma-separated method arguments")
	cmd.Flags().StringVar(&value, "value", "", "ETH value to send (e.g. '0.01')")
	cmd.Flags().Int64Var(&chainID, "chain-id", 0, "Chain ID (default: from config)")
	cmd.Flags().BoolVar(&asJSON, "output", false, "Output as JSON")

	_ = cmd.MarkFlagRequired("address")
	_ = cmd.MarkFlagRequired("abi")
	_ = cmd.MarkFlagRequired("method")

	return cmd
}
