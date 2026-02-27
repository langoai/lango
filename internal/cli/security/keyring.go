package security

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
		Short: "Store the master passphrase in a secure hardware backend",
		Long: `Store the master passphrase using the best available secure hardware backend:

  - macOS with Touch ID:  Keychain with biometric access control
  - Linux with TPM 2.0:   TPM-sealed blob (~/.lango/tpm/)

If no secure hardware backend is available, this command will refuse to store
the passphrase to avoid exposing it to same-UID attacks via plain OS keyring.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			secureProvider, tier := keyring.DetectSecureProvider()
			if secureProvider == nil {
				return fmt.Errorf(
					"no secure hardware backend available (security tier: %s)\n"+
						"Use a keyfile (LANGO_PASSPHRASE_FILE) or interactive prompt instead",
					tier.String(),
				)
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

			pass, err := prompt.Passphrase("Enter passphrase to store: ")
			if err != nil {
				return fmt.Errorf("read passphrase: %w", err)
			}

			if err := secureProvider.Set(keyring.Service, keyring.KeyMasterPassphrase, pass); err != nil {
				return fmt.Errorf("store passphrase: %w", err)
			}

			fmt.Printf("Passphrase stored (security tier: %s).\n", tier.String())
			return nil
		},
	}
}

func newKeyringClearCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Remove the master passphrase from all storage backends",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				if !prompt.IsInteractive() {
					return fmt.Errorf("use --force for non-interactive deletion")
				}
				ok, err := prompt.Confirm("Remove passphrase from all keyring backends?")
				if err != nil {
					return err
				}
				if !ok {
					fmt.Println("Aborted.")
					return nil
				}
			}

			var cleared int

			// 1. Try OS keyring (go-keyring) — covers non-biometric Keychain items.
			if status := keyring.IsAvailable(); status.Available {
				provider := keyring.NewOSProvider()
				if err := provider.Delete(keyring.Service, keyring.KeyMasterPassphrase); err == nil {
					fmt.Println("Removed passphrase from OS keyring.")
					cleared++
				} else if !errors.Is(err, keyring.ErrNotFound) {
					fmt.Fprintf(os.Stderr, "warning: OS keyring delete: %v\n", err)
				}
			}

			// 2. Try secure provider — only for non-macOS where the secure backend
			// is distinct from the OS keyring (e.g., TPM sealed blobs on Linux).
			// On macOS, OSProvider.Delete already handles Keychain items (both
			// biometric and non-biometric share the same service/account key).
			if runtime.GOOS != "darwin" {
				if secureProvider, _ := keyring.DetectSecureProvider(); secureProvider != nil {
					if err := secureProvider.Delete(keyring.Service, keyring.KeyMasterPassphrase); err == nil {
						fmt.Println("Removed passphrase from secure provider.")
						cleared++
					} else if !errors.Is(err, keyring.ErrNotFound) {
						fmt.Fprintf(os.Stderr, "warning: secure provider delete: %v\n", err)
					}
				}
			}

			// 3. Remove TPM sealed blob files if they exist (belt-and-suspenders).
			home, err := os.UserHomeDir()
			if err == nil {
				tpmDir := filepath.Join(home, ".lango", "tpm")
				blobPath := filepath.Join(tpmDir, keyring.Service+"_"+keyring.KeyMasterPassphrase+".sealed")
				if err := os.Remove(blobPath); err == nil {
					fmt.Println("Removed TPM sealed blob file.")
					cleared++
				}
			}

			if cleared == 0 {
				fmt.Println("No stored passphrase found in any backend.")
			}

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
		Short: "Show keyring availability, security tier, and stored passphrase status",
		RunE: func(cmd *cobra.Command, args []string) error {
			status := keyring.IsAvailable()

			// Check for stored passphrase in OS keyring.
			hasPassphraseOS := false
			if status.Available {
				provider := keyring.NewOSProvider()
				_, err := provider.Get(keyring.Service, keyring.KeyMasterPassphrase)
				hasPassphraseOS = err == nil
			}

			// Check for stored passphrase in secure provider using HasKey
			// (avoids triggering Touch ID just for a status check).
			hasPassphraseSecure := false
			secureProvider, tier := keyring.DetectSecureProvider()
			if secureProvider != nil {
				if checker, ok := secureProvider.(keyring.KeyChecker); ok {
					hasPassphraseSecure = checker.HasKey(keyring.Service, keyring.KeyMasterPassphrase)
				}
			}

			type statusOutput struct {
				Available           bool   `json:"available"`
				Backend             string `json:"backend,omitempty"`
				SecurityTier        string `json:"security_tier"`
				Error               string `json:"error,omitempty"`
				HasPassphraseOS     bool   `json:"has_passphrase_os"`
				HasPassphraseSecure bool   `json:"has_passphrase_secure"`
			}

			out := statusOutput{
				Available:           status.Available,
				Backend:             status.Backend,
				SecurityTier:        tier.String(),
				Error:               status.Error,
				HasPassphraseOS:     hasPassphraseOS,
				HasPassphraseSecure: hasPassphraseSecure,
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Println("Keyring Status")
			fmt.Printf("  Available:               %v\n", out.Available)
			if out.Backend != "" {
				fmt.Printf("  Backend:                 %s\n", out.Backend)
			}
			fmt.Printf("  Security Tier:           %s\n", out.SecurityTier)
			if out.Error != "" {
				fmt.Printf("  Error:                   %s\n", out.Error)
			}
			fmt.Printf("  Has Passphrase (OS):     %v\n", out.HasPassphraseOS)
			fmt.Printf("  Has Passphrase (Secure): %v\n", out.HasPassphraseSecure)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
