package smartaccount

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	sa "github.com/langoai/lango/internal/smartaccount"
)

type listedModuleEntry struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Address string `json:"address"`
	Version string `json:"version"`
}

var installSmartAccountModule = func(bootLoader BootLoader, modType sa.ModuleType, addr common.Address) (string, func(), error) {
	boot, err := bootLoader()
	if err != nil {
		return "", nil, fmt.Errorf("bootstrap: %w", err)
	}

	deps, err := initSmartAccountDeps(boot)
	if err != nil {
		boot.Close()
		return "", nil, err
	}

	ctx := context.Background()
	txHash, err := deps.manager.InstallModule(ctx, modType, addr, []byte{})
	if err != nil {
		deps.cleanup()
		boot.Close()
		return "", nil, fmt.Errorf("install module: %w", err)
	}

	return txHash, func() {
		deps.cleanup()
		boot.Close()
	}, nil
}

var loadModuleListEntries = func(bootLoader BootLoader) ([]listedModuleEntry, func(), error) {
	boot, err := bootLoader()
	if err != nil {
		return nil, nil, fmt.Errorf("bootstrap: %w", err)
	}

	deps, err := initSmartAccountDeps(boot)
	if err != nil {
		boot.Close()
		return nil, nil, err
	}

	modules := deps.moduleRegistry.List()
	entries := make([]listedModuleEntry, 0, len(modules))
	for _, m := range modules {
		entries = append(entries, listedModuleEntry{
			Name:    m.Name,
			Type:    m.Type.String(),
			Address: m.Address.Hex(),
			Version: m.Version,
		})
	}

	return entries, func() {
		deps.cleanup()
		boot.Close()
	}, nil
}

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
	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List registered ERC-7579 modules",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveTableOrJSONOutput(cmd)
			if err != nil {
				return err
			}
			entries, cleanup, err := loadModuleListEntries(bootLoader)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), entries)
			}

			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No modules registered.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTYPE\tADDRESS\tVERSION")
			for _, m := range entries {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Name, m.Type, m.Address, m.Version)
			}
			return w.Flush()
		},
	}

	cmd.Flags().String("output", "table", "output format (table|json)")
	return cmd
}

func moduleInstallCmd(bootLoader BootLoader) *cobra.Command {
	var moduleType string

	cmd := &cobra.Command{
		Use:   "install <module-address>",
		Short: "Install an ERC-7579 module on the smart account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			txHash, cleanup, err := installSmartAccountModule(bootLoader, modType, addr)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Module installed successfully.")
			fmt.Fprintf(out, "  Address:  %s\n", addr.Hex())
			fmt.Fprintf(out, "  Type:     %s\n", modType.String())
			fmt.Fprintf(out, "  Tx Hash:  %s\n", txHash)

			return nil
		},
	}

	cmd.Flags().StringVar(&moduleType, "type", "validator", "module type (validator|executor|fallback|hook)")
	return cmd
}
