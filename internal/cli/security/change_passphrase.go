package security

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/keyring"
	"github.com/langoai/lango/internal/security"
)

// newChangePassphraseCmd creates `lango security change-passphrase`, the envelope-aware
// replacement for `migrate-passphrase`. Unlike migrate-passphrase, this command does
// NOT re-encrypt any data or rekey the SQLCipher database — it re-wraps the Master Key
// in-place, so the operation is O(1) in the amount of stored data.
func newChangePassphraseCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "change-passphrase",
		Short: "Change the passphrase by re-wrapping the Master Key (no data re-encryption)",
		Long: `Change the passphrase that protects the Master Key.

This command re-wraps the existing Master Key with a new passphrase-derived
KEK. Because the MK itself does not change, stored secrets, configuration
profiles, and the SQLCipher database key all remain valid — no data is
re-encrypted and no PRAGMA rekey is issued.

Recovery mnemonic slots (if present) are unchanged.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			if !prompt.IsInteractive() {
				return fmt.Errorf("this command requires an interactive terminal")
			}

			provider, ok := boot.Crypto.(*security.LocalCryptoProvider)
			if !ok {
				return fmt.Errorf("change-passphrase is only available for the local crypto provider")
			}
			envelope := provider.Envelope()
			if envelope == nil {
				return fmt.Errorf("envelope not found — this installation still uses the legacy format. " +
					"Run `lango security migrate-passphrase` or upgrade the install first")
			}

			// We need the MK to re-wrap. The provider installed it as keys["local"];
			// verify the current passphrase and unwrap a fresh copy from the envelope
			// rather than reaching into the provider's internal state.
			currentPass, err := prompt.Passphrase("Enter CURRENT passphrase: ")
			if err != nil {
				return fmt.Errorf("read current passphrase: %w", err)
			}
			mk, _, err := envelope.UnwrapFromPassphrase(currentPass)
			if err != nil {
				return fmt.Errorf("current passphrase is incorrect")
			}
			defer security.ZeroBytes(mk)

			newPass, err := prompt.PassphraseConfirm("Enter NEW passphrase: ", "Confirm NEW passphrase: ")
			if err != nil {
				return err
			}
			if len(newPass) < 8 {
				return fmt.Errorf("passphrase must be at least 8 characters")
			}

			if err := envelope.ChangePassphraseSlot(mk, newPass); err != nil {
				return fmt.Errorf("rotate passphrase slot: %w", err)
			}

			if err := security.StoreEnvelopeFile(boot.LangoDir, envelope); err != nil {
				return fmt.Errorf("persist envelope: %w", err)
			}

			// Sync stored credentials so next headless/keyring bootstrap works.
			keyfilePath := filepath.Join(boot.LangoDir, "keyfile")
			if _, statErr := os.Stat(keyfilePath); statErr == nil {
				if writeErr := os.WriteFile(keyfilePath, []byte(newPass), 0600); writeErr != nil {
					fmt.Fprintf(os.Stderr, "warning: update keyfile: %v\n", writeErr)
				} else {
					fmt.Fprintln(os.Stderr, "Keyfile updated with new passphrase.")
				}
			}
			if secureProvider, _ := keyring.DetectSecureProvider(); secureProvider != nil {
				// Always attempt keyring update. This command requires an interactive
				// terminal, so the biometric prompt is visible to the user. Skipping
				// the update risks leaving a stale entry that breaks headless bootstrap.
				if setErr := secureProvider.Set(keyring.Service, keyring.KeyMasterPassphrase, newPass); setErr != nil {
					fmt.Fprintf(os.Stderr, "warning: keyring update failed: %v\n", setErr)
					fmt.Fprintf(os.Stderr, "  If a stale passphrase is stored, next headless bootstrap may fail.\n")
					fmt.Fprintf(os.Stderr, "  Fix: run `lango security keyring set` or clear the keyring entry manually.\n")
				} else {
					fmt.Fprintln(os.Stderr, "Keyring updated with new passphrase.")
				}
			}

			fmt.Println("Passphrase changed. No data was re-encrypted.")
			return nil
		},
	}
}
