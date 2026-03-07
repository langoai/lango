package contract

import (
	"encoding/json"
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
		asJSON  bool
	)

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Parse and validate a contract ABI from file",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"address": address,
					"chainId": chainID,
					"methods": methodCount,
					"events":  eventCount,
					"status":  "loaded",
				})
			}

			fmt.Printf("ABI Loaded\n")
			fmt.Printf("  Address:  %s\n", address)
			fmt.Printf("  Chain ID: %d\n", chainID)
			fmt.Printf("  Methods:  %d\n", methodCount)
			fmt.Printf("  Events:   %d\n", eventCount)

			return nil
		},
	}

	cmd.Flags().StringVar(&address, "address", "", "Contract address (0x...)")
	cmd.Flags().StringVar(&file, "file", "", "Path to ABI JSON file")
	cmd.Flags().Int64Var(&chainID, "chain-id", 0, "Chain ID (default: from config)")
	cmd.Flags().BoolVar(&asJSON, "output", false, "Output as JSON")

	_ = cmd.MarkFlagRequired("address")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
