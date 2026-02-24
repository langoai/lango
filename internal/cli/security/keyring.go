package security

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/keyring"
)

func newKeyringCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keyring",
		Short: "Manage OS keyring passphrase storage",
	}

	cmd.AddCommand(newKeyringStoreCmd(bootLoader))
	cmd.AddCommand(newKeyringClearCmd())
	cmd.AddCommand(newKeyringStatusCmd())

	return cmd
}

func newKeyringStoreCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "store",
		Short: "Store the master passphrase in the OS keyring",
		Long: `Store the master passphrase in the OS keyring (macOS Keychain,
Linux secret-service, or Windows Credential Manager).

This allows lango to unlock automatically without a keyfile or interactive prompt.
The passphrase is verified against the existing crypto state before storing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			status := keyring.IsAvailable()
			if !status.Available {
				return fmt.Errorf("OS keyring not available: %s", status.Error)
			}

			// Bootstrap to verify the passphrase is correct.
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			if !prompt.IsInteractive() {
				return fmt.Errorf("this command requires an interactive terminal")
			}

			pass, err := prompt.Passphrase("Enter passphrase to store in keyring: ")
			if err != nil {
				return fmt.Errorf("read passphrase: %w", err)
			}

			provider := keyring.NewOSProvider()
			if err := provider.Set(keyring.Service, keyring.KeyMasterPassphrase, pass); err != nil {
				return fmt.Errorf("store passphrase in keyring: %w", err)
			}

			fmt.Printf("Passphrase stored in OS keyring (%s).\n", status.Backend)
			return nil
		},
	}
}

func newKeyringClearCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Remove the master passphrase from the OS keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			status := keyring.IsAvailable()
			if !status.Available {
				return fmt.Errorf("OS keyring not available: %s", status.Error)
			}

			if !force {
				if !prompt.IsInteractive() {
					return fmt.Errorf("use --force for non-interactive deletion")
				}
				fmt.Print("Remove passphrase from OS keyring? [y/N] ")
				var answer string
				fmt.Scanln(&answer)
				if answer != "y" && answer != "Y" && answer != "yes" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			provider := keyring.NewOSProvider()
			if err := provider.Delete(keyring.Service, keyring.KeyMasterPassphrase); err != nil {
				if errors.Is(err, keyring.ErrNotFound) {
					fmt.Println("No passphrase stored in keyring.")
					return nil
				}
				return fmt.Errorf("remove passphrase from keyring: %w", err)
			}

			fmt.Println("Passphrase removed from OS keyring.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}

func newKeyringStatusCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show OS keyring availability and stored passphrase status",
		RunE: func(cmd *cobra.Command, args []string) error {
			status := keyring.IsAvailable()

			hasPassphrase := false
			if status.Available {
				provider := keyring.NewOSProvider()
				_, err := provider.Get(keyring.Service, keyring.KeyMasterPassphrase)
				hasPassphrase = err == nil
			}

			type statusOutput struct {
				Available     bool   `json:"available"`
				Backend       string `json:"backend,omitempty"`
				Error         string `json:"error,omitempty"`
				HasPassphrase bool   `json:"has_passphrase"`
			}

			out := statusOutput{
				Available:     status.Available,
				Backend:       status.Backend,
				Error:         status.Error,
				HasPassphrase: hasPassphrase,
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Println("OS Keyring Status")
			fmt.Printf("  Available:      %v\n", out.Available)
			if out.Backend != "" {
				fmt.Printf("  Backend:        %s\n", out.Backend)
			}
			if out.Error != "" {
				fmt.Printf("  Error:          %s\n", out.Error)
			}
			fmt.Printf("  Has Passphrase: %v\n", out.HasPassphrase)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
