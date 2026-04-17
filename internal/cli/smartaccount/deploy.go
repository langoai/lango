package smartaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func deployCmd(bootLoader BootLoader) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new Safe smart account with ERC-7579 adapter",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			deps, err := initSmartAccountDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			ctx := context.Background()
			info, err := deps.manager.GetOrDeploy(ctx)
			if err != nil {
				return fmt.Errorf("deploy account: %w", err)
			}

			type deployResult struct {
				Address    string `json:"address"`
				IsDeployed bool   `json:"isDeployed"`
				Owner      string `json:"ownerAddress"`
				ChainID    int64  `json:"chainId"`
				EntryPoint string `json:"entryPoint"`
				Modules    int    `json:"moduleCount"`
			}

			result := deployResult{
				Address:    info.Address.Hex(),
				IsDeployed: info.IsDeployed,
				Owner:      info.OwnerAddress.Hex(),
				ChainID:    info.ChainID,
				EntryPoint: info.EntryPoint.Hex(),
				Modules:    len(info.Modules),
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(result, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println("Smart Account Deployed")
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "  Address:\t%s\n", result.Address)
			fmt.Fprintf(w, "  Deployed:\t%v\n", result.IsDeployed)
			fmt.Fprintf(w, "  Owner:\t%s\n", result.Owner)
			fmt.Fprintf(w, "  Chain ID:\t%d\n", result.ChainID)
			fmt.Fprintf(w, "  Entry Point:\t%s\n", result.EntryPoint)
			fmt.Fprintf(w, "  Modules:\t%d\n", result.Modules)
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	return cmd
}
