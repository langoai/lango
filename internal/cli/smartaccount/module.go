package smartaccount

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func moduleCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module",
		Short: "Manage ERC-7579 modules",
		Long: `Manage ERC-7579 modules for smart account extensibility.

Examples:
  lango account module list
  lango account module install <module-name>`,
	}

	cmd.AddCommand(moduleListCmd(bootLoader))
	cmd.AddCommand(moduleInstallCmd(bootLoader))

	return cmd
}

func moduleListCmd(bootLoader BootLoader) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured ERC-7579 modules",
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

			type moduleEntry struct {
				Name    string `json:"name"`
				Type    string `json:"type"`
				Address string `json:"address"`
			}

			modules := make([]moduleEntry, 0, 3)
			if cfg.SmartAccount.Modules.SessionValidatorAddress != "" {
				modules = append(modules, moduleEntry{
					Name:    "SessionValidator",
					Type:    "validator",
					Address: cfg.SmartAccount.Modules.SessionValidatorAddress,
				})
			}
			if cfg.SmartAccount.Modules.SpendingHookAddress != "" {
				modules = append(modules, moduleEntry{
					Name:    "SpendingHook",
					Type:    "hook",
					Address: cfg.SmartAccount.Modules.SpendingHookAddress,
				})
			}
			if cfg.SmartAccount.Modules.EscrowExecutorAddress != "" {
				modules = append(modules, moduleEntry{
					Name:    "EscrowExecutor",
					Type:    "executor",
					Address: cfg.SmartAccount.Modules.EscrowExecutorAddress,
				})
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(modules, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			if len(modules) == 0 {
				fmt.Println("No modules configured.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTYPE\tADDRESS")
			for _, m := range modules {
				fmt.Fprintf(w, "%s\t%s\t%s\n", m.Name, m.Type, m.Address)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	return cmd
}

func moduleInstallCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <module-name>",
		Short: "Install an ERC-7579 module",
		Args:  cobra.ExactArgs(1),
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

			moduleName := args[0]
			fmt.Printf("Installing module: %s\n", moduleName)
			fmt.Println()
			fmt.Println("Note: Module installation requires a running server (lango serve).")
			fmt.Println("Use the 'smart_account_install_module' agent tool for actual installation.")

			return nil
		},
	}

	return cmd
}
