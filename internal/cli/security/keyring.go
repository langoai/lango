package security

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/keyring"
)

func newKeyringCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keyring",
		Short: "Manage hardware keyring passphrase storage (Touch ID / TPM)",
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
			defer boot.Close()

			// Check if passphrase is already stored in the secure provider.
			if checker, ok := secureProvider.(keyring.KeyChecker); ok {
				if checker.HasKey(keyring.Service, keyring.KeyMasterPassphrase) {
					fmt.Println("Passphrase is already stored in the secure keyring.")
					fmt.Println("  Next launch will load it automatically.")
					return nil
				}
			}

			if !prompt.IsInteractive() {
				return fmt.Errorf("this command requires an interactive terminal")
			}

			pass, err := prompt.Passphrase("Enter passphrase to store: ")
			if err != nil {
				return fmt.Errorf("read passphrase: %w", err)
			}

			if err := secureProvider.Set(keyring.Service, keyring.KeyMasterPassphrase, pass); err != nil {
				if errors.Is(err, keyring.ErrEntitlement) {
					return fmt.Errorf("biometric storage unavailable (binary not codesigned)\n" +
						"  Tip: codesign the binary: make codesign\n" +
						"  Note: also ensure device passcode is set (required for biometric Keychain)")
				}
				return fmt.Errorf("store passphrase: %w", err)
			}

			fmt.Printf("Passphrase stored with %s protection.\n", tier.String())
			fmt.Println("  Next launch will load it automatically.")
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

			// 1. Try secure hardware provider (biometric / TPM).
			if secureProvider, _ := keyring.DetectSecureProvider(); secureProvider != nil {
				if err := secureProvider.Delete(keyring.Service, keyring.KeyMasterPassphrase); err == nil {
					fmt.Println("Removed passphrase from secure provider.")
					cleared++
				} else if !errors.Is(err, keyring.ErrNotFound) {
					fmt.Fprintf(os.Stderr, "warning: secure provider delete: %v\n", err)
				}
			}

			// 2. Remove TPM sealed blob files if they exist (belt-and-suspenders).
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
			// Detect hardware-backed secure provider (biometric / TPM).
			secureProvider, tier := keyring.DetectSecureProvider()
			available := secureProvider != nil

			// Check for stored passphrase using HasKey (avoids triggering Touch ID).
			hasPassphrase := false
			if secureProvider != nil {
				if checker, ok := secureProvider.(keyring.KeyChecker); ok {
					hasPassphrase = checker.HasKey(keyring.Service, keyring.KeyMasterPassphrase)
				}
			}

			type statusOutput struct {
				Available     bool   `json:"available"`
				SecurityTier  string `json:"security_tier"`
				HasPassphrase bool   `json:"has_passphrase"`
			}

			out := statusOutput{
				Available:     available,
				SecurityTier:  tier.String(),
				HasPassphrase: hasPassphrase,
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Println("Hardware Keyring Status")
			fmt.Printf("  Available:       %v\n", out.Available)
			fmt.Printf("  Security Tier:   %s\n", out.SecurityTier)
			fmt.Printf("  Has Passphrase:  %v\n", out.HasPassphrase)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
