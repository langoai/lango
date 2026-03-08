package smartaccount

import (
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

			cfg := boot.Config

			type moduleAddresses struct {
				SessionValidator string `json:"sessionValidator"`
				SpendingHook     string `json:"spendingHook"`
				EscrowExecutor   string `json:"escrowExecutor"`
			}

			type sessionConfig struct {
				MaxDuration string `json:"maxDuration"`
				MaxKeys     int    `json:"maxKeys"`
				DefaultGas  uint64 `json:"defaultGasLimit"`
			}

			type accountInfo struct {
				Enabled    bool            `json:"enabled"`
				Factory    string          `json:"factory"`
				EntryPoint string          `json:"entryPoint"`
				Safe7579   string          `json:"safe7579"`
				Fallback   string          `json:"fallbackHandler"`
				Bundler    string          `json:"bundler"`
				Session    sessionConfig   `json:"session"`
				Modules    moduleAddresses `json:"modules"`
			}

			info := accountInfo{
				Enabled:    cfg.SmartAccount.Enabled,
				Factory:    cfg.SmartAccount.FactoryAddress,
				EntryPoint: cfg.SmartAccount.EntryPointAddress,
				Safe7579:   cfg.SmartAccount.Safe7579Address,
				Fallback:   cfg.SmartAccount.FallbackHandler,
				Bundler:    cfg.SmartAccount.BundlerURL,
				Session: sessionConfig{
					MaxDuration: cfg.SmartAccount.Session.MaxDuration.String(),
					MaxKeys:     cfg.SmartAccount.Session.MaxActiveKeys,
					DefaultGas:  cfg.SmartAccount.Session.DefaultGasLimit,
				},
				Modules: moduleAddresses{
					SessionValidator: cfg.SmartAccount.Modules.SessionValidatorAddress,
					SpendingHook:     cfg.SmartAccount.Modules.SpendingHookAddress,
					EscrowExecutor:   cfg.SmartAccount.Modules.EscrowExecutorAddress,
				},
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(info, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "Smart Account Configuration")
			fmt.Fprintln(w, "===========================")
			fmt.Fprintf(w, "Enabled:\t%v\n", info.Enabled)
			fmt.Fprintf(w, "Factory:\t%s\n", info.Factory)
			fmt.Fprintf(w, "Entry Point:\t%s\n", info.EntryPoint)
			fmt.Fprintf(w, "Safe7579:\t%s\n", info.Safe7579)
			fmt.Fprintf(w, "Fallback Handler:\t%s\n", info.Fallback)
			fmt.Fprintf(w, "Bundler URL:\t%s\n", info.Bundler)
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Session Configuration")
			fmt.Fprintln(w, "---------------------")
			fmt.Fprintf(w, "Max Duration:\t%s\n", info.Session.MaxDuration)
			fmt.Fprintf(w, "Max Active Keys:\t%d\n", info.Session.MaxKeys)
			fmt.Fprintf(w, "Default Gas Limit:\t%d\n", info.Session.DefaultGas)
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Module Addresses")
			fmt.Fprintln(w, "----------------")
			fmt.Fprintf(w, "Session Validator:\t%s\n", info.Modules.SessionValidator)
			fmt.Fprintf(w, "Spending Hook:\t%s\n", info.Modules.SpendingHook)
			fmt.Fprintf(w, "Escrow Executor:\t%s\n", info.Modules.EscrowExecutor)

			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	return cmd
}
