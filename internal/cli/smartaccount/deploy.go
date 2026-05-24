package smartaccount

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type deployAccountResult struct {
	Address    string `json:"address"`
	IsDeployed bool   `json:"isDeployed"`
	Owner      string `json:"ownerAddress"`
	ChainID    int64  `json:"chainId"`
	EntryPoint string `json:"entryPoint"`
	Modules    int    `json:"moduleCount"`
}

var loadDeployAccountResult = func(bootLoader BootLoader) (deployAccountResult, func(), error) {
	boot, err := bootLoader()
	if err != nil {
		return deployAccountResult{}, nil, fmt.Errorf("bootstrap: %w", err)
	}

	deps, err := initSmartAccountDeps(boot)
	if err != nil {
		boot.Close()
		return deployAccountResult{}, nil, err
	}

	ctx := context.Background()
	info, err := deps.manager.GetOrDeploy(ctx)
	if err != nil {
		deps.cleanup()
		boot.Close()
		return deployAccountResult{}, nil, fmt.Errorf("deploy account: %w", err)
	}

	result := deployAccountResult{
		Address:    info.Address.Hex(),
		IsDeployed: info.IsDeployed,
		Owner:      info.OwnerAddress.Hex(),
		ChainID:    info.ChainID,
		EntryPoint: info.EntryPoint.Hex(),
		Modules:    len(info.Modules),
	}

	return result, func() {
		deps.cleanup()
		boot.Close()
	}, nil
}

func deployCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "deploy",
		Short:         "Deploy a new Safe smart account with ERC-7579 adapter",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveTableOrJSONOutput(cmd)
			if err != nil {
				return err
			}
			result, cleanup, err := loadDeployAccountResult(bootLoader)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), result)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Smart Account Deployed")
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "  Address:\t%s\n", result.Address)
			fmt.Fprintf(w, "  Deployed:\t%v\n", result.IsDeployed)
			fmt.Fprintf(w, "  Owner:\t%s\n", result.Owner)
			fmt.Fprintf(w, "  Chain ID:\t%d\n", result.ChainID)
			fmt.Fprintf(w, "  Entry Point:\t%s\n", result.EntryPoint)
			fmt.Fprintf(w, "  Modules:\t%d\n", result.Modules)
			return w.Flush()
		},
	}

	cmd.Flags().String("output", "table", "output format (table|json)")
	return cmd
}
