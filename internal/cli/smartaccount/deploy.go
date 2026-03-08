package smartaccount

import (
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
			defer boot.DBClient.Close()

			cfg := boot.Config
			if !cfg.SmartAccount.Enabled {
				return fmt.Errorf("smart account not enabled in config")
			}

			type deployInfo struct {
				Factory    string `json:"factory"`
				EntryPoint string `json:"entryPoint"`
				Safe7579   string `json:"safe7579"`
				Bundler    string `json:"bundler"`
			}

			info := deployInfo{
				Factory:    cfg.SmartAccount.FactoryAddress,
				EntryPoint: cfg.SmartAccount.EntryPointAddress,
				Safe7579:   cfg.SmartAccount.Safe7579Address,
				Bundler:    cfg.SmartAccount.BundlerURL,
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(info, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println("Deploying Smart Account...")
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "  Factory:\t%s\n", info.Factory)
			fmt.Fprintf(w, "  Entry Point:\t%s\n", info.EntryPoint)
			fmt.Fprintf(w, "  Safe7579:\t%s\n", info.Safe7579)
			fmt.Fprintf(w, "  Bundler:\t%s\n", info.Bundler)
			if flushErr := w.Flush(); flushErr != nil {
				return fmt.Errorf("flush output: %w", flushErr)
			}

			fmt.Println()
			fmt.Println("Note: Full deployment requires a running server (lango serve).")
			fmt.Println("Use the 'smart_account_deploy' agent tool for actual deployment.")

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	return cmd
}
