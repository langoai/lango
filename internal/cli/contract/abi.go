package contract

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	contractpkg "github.com/langoai/lango/internal/contract"
)

func newABICmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abi",
		Short: "ABI management commands",
	}

	cmd.AddCommand(newABILoadCmd(cfgLoader))

	return cmd
}

func newABILoadCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		address string
		file    string
		chainID int64
		output  string
	)

	cmd := &cobra.Command{
		Use:           "load",
		Short:         "Parse and validate a contract ABI from file",
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

			abiJSON, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("read ABI file %q: %w", file, err)
			}

			if chainID == 0 {
				chainID = cfg.Payment.Network.ChainID
			}

			cache := contractpkg.NewABICache()
			parsed, err := cache.GetOrParse(chainID, common.HexToAddress(address), string(abiJSON))
			if err != nil {
				return fmt.Errorf("parse ABI: %w", err)
			}

			methodCount := len(parsed.Methods)
			eventCount := len(parsed.Events)

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), map[string]interface{}{
					"address": address,
					"chainId": chainID,
					"methods": methodCount,
					"events":  eventCount,
					"status":  "loaded",
				})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "ABI Loaded")
			fmt.Fprintf(cmd.OutOrStdout(), "  Address:  %s\n", address)
			fmt.Fprintf(cmd.OutOrStdout(), "  Chain ID: %d\n", chainID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Methods:  %d\n", methodCount)
			fmt.Fprintf(cmd.OutOrStdout(), "  Events:   %d\n", eventCount)

			return nil
		},
	}

	cmd.Flags().StringVar(&address, "address", "", "Contract address (0x...)")
	cmd.Flags().StringVar(&file, "file", "", "Path to ABI JSON file")
	cmd.Flags().Int64Var(&chainID, "chain-id", 0, "Chain ID (default: from config)")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")

	_ = cmd.MarkFlagRequired("address")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
