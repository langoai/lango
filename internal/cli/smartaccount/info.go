package smartaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func infoCmd(bootLoader BootLoader) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show smart account configuration and status",
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

			ctx := context.Background()
			info, err := deps.manager.Info(ctx)
			if err != nil {
				return fmt.Errorf("get account info: %w", err)
			}

			type moduleEntry struct {
				Name    string `json:"name"`
				Type    string `json:"type"`
				Address string `json:"address"`
			}

			type accountInfo struct {
				Address    string        `json:"address"`
				IsDeployed bool          `json:"isDeployed"`
				Owner      string        `json:"ownerAddress"`
				ChainID    int64         `json:"chainId"`
				EntryPoint string        `json:"entryPoint"`
				Modules    []moduleEntry `json:"modules"`
				Paymaster  bool          `json:"paymasterEnabled"`
			}

			modules := make([]moduleEntry, 0, len(info.Modules))
			for _, m := range info.Modules {
				modules = append(modules, moduleEntry{
					Name:    m.Name,
					Type:    m.Type.String(),
					Address: m.Address.Hex(),
				})
			}

			result := accountInfo{
				Address:    info.Address.Hex(),
				IsDeployed: info.IsDeployed,
				Owner:      info.OwnerAddress.Hex(),
				ChainID:    info.ChainID,
				EntryPoint: info.EntryPoint.Hex(),
				Modules:    modules,
				Paymaster:  deps.paymasterProv != nil,
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(result, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
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

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	return cmd
}
