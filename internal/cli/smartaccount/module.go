package smartaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	sa "github.com/langoai/lango/internal/smartaccount"
)

func moduleCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module",
		Short: "Manage ERC-7579 modules",
		Long: `Manage ERC-7579 modules for smart account extensibility.

Examples:
  lango account module list
  lango account module install <module-address> --type validator`,
	}

	cmd.AddCommand(moduleListCmd(bootLoader))
	cmd.AddCommand(moduleInstallCmd(bootLoader))

	return cmd
}

func moduleListCmd(bootLoader BootLoader) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered ERC-7579 modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			deps, err := initSmartAccountDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			modules := deps.moduleRegistry.List()

			type moduleEntry struct {
				Name    string `json:"name"`
				Type    string `json:"type"`
				Address string `json:"address"`
				Version string `json:"version"`
			}

			entries := make([]moduleEntry, 0, len(modules))
			for _, m := range modules {
				entries = append(entries, moduleEntry{
					Name:    m.Name,
					Type:    m.Type.String(),
					Address: m.Address.Hex(),
					Version: m.Version,
				})
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(entries, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			if len(entries) == 0 {
				fmt.Println("No modules registered.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTYPE\tADDRESS\tVERSION")
			for _, m := range entries {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Name, m.Type, m.Address, m.Version)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	return cmd
}

func moduleInstallCmd(bootLoader BootLoader) *cobra.Command {
	var moduleType string

	cmd := &cobra.Command{
		Use:   "install <module-address>",
		Short: "Install an ERC-7579 module on the smart account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			deps, err := initSmartAccountDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			addrStr := args[0]
			if !common.IsHexAddress(addrStr) {
				return fmt.Errorf("invalid module address: %s", addrStr)
			}
			addr := common.HexToAddress(addrStr)

			// Parse module type.
			var modType sa.ModuleType
			switch moduleType {
			case "validator":
				modType = sa.ModuleTypeValidator
			case "executor":
				modType = sa.ModuleTypeExecutor
			case "fallback":
				modType = sa.ModuleTypeFallback
			case "hook":
				modType = sa.ModuleTypeHook
			default:
				return fmt.Errorf("unknown module type %q (use: validator, executor, fallback, hook)", moduleType)
			}

			ctx := context.Background()
			txHash, err := deps.manager.InstallModule(ctx, modType, addr, []byte{})
			if err != nil {
				return fmt.Errorf("install module: %w", err)
			}

			fmt.Printf("Module installed successfully.\n")
			fmt.Printf("  Address:  %s\n", addr.Hex())
			fmt.Printf("  Type:     %s\n", modType.String())
			fmt.Printf("  Tx Hash:  %s\n", txHash)

			return nil
		},
	}

	cmd.Flags().StringVar(&moduleType, "type", "validator", "module type (validator|executor|fallback|hook)")
	return cmd
}
