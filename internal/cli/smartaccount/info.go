package smartaccount

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type infoModuleEntry struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Address string `json:"address"`
}

type infoAccountResult struct {
	Address    string            `json:"address"`
	IsDeployed bool              `json:"isDeployed"`
	Owner      string            `json:"ownerAddress"`
	ChainID    int64             `json:"chainId"`
	EntryPoint string            `json:"entryPoint"`
	Modules    []infoModuleEntry `json:"modules"`
	Paymaster  bool              `json:"paymasterEnabled"`
}

var loadInfoAccountResult = func(bootLoader BootLoader) (infoAccountResult, func(), error) {
	boot, err := bootLoader()
	if err != nil {
		return infoAccountResult{}, nil, fmt.Errorf("bootstrap: %w", err)
	}

	deps, err := initSmartAccountDeps(boot)
	if err != nil {
		boot.Close()
		return infoAccountResult{}, nil, err
	}

	ctx := context.Background()
	info, err := deps.manager.Info(ctx)
	if err != nil {
		deps.cleanup()
		boot.Close()
		return infoAccountResult{}, nil, fmt.Errorf("get account info: %w", err)
	}

	modules := make([]infoModuleEntry, 0, len(info.Modules))
	for _, m := range info.Modules {
		modules = append(modules, infoModuleEntry{
			Name:    m.Name,
			Type:    m.Type.String(),
			Address: m.Address.Hex(),
		})
	}

	result := infoAccountResult{
		Address:    info.Address.Hex(),
		IsDeployed: info.IsDeployed,
		Owner:      info.OwnerAddress.Hex(),
		ChainID:    info.ChainID,
		EntryPoint: info.EntryPoint.Hex(),
		Modules:    modules,
		Paymaster:  deps.paymasterProv != nil,
	}

	return result, func() {
		deps.cleanup()
		boot.Close()
	}, nil
}

func infoCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "info",
		Short:         "Show smart account configuration and status",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveTableOrJSONOutput(cmd)
			if err != nil {
				return err
			}
			result, cleanup, err := loadInfoAccountResult(bootLoader)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), result)
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "Smart Account Info")
			fmt.Fprintln(w, "==================")
			fmt.Fprintf(w, "Address:\t%s\n", result.Address)
			fmt.Fprintf(w, "Deployed:\t%v\n", result.IsDeployed)
			fmt.Fprintf(w, "Owner:\t%s\n", result.Owner)
			fmt.Fprintf(w, "Chain ID:\t%d\n", result.ChainID)
			fmt.Fprintf(w, "Entry Point:\t%s\n", result.EntryPoint)
			fmt.Fprintf(w, "Paymaster:\t%v\n", result.Paymaster)
			fmt.Fprintln(w)

			if len(result.Modules) > 0 {
				fmt.Fprintln(w, "Installed Modules")
				fmt.Fprintln(w, "-----------------")
				fmt.Fprintln(w, "NAME\tTYPE\tADDRESS")
				for _, m := range result.Modules {
					fmt.Fprintf(w, "%s\t%s\t%s\n", m.Name, m.Type, m.Address)
				}
			} else {
				fmt.Fprintln(w, "No modules installed.")
			}

			return w.Flush()
		},
	}

	cmd.Flags().String("output", "table", "output format (table|json)")
	return cmd
}
