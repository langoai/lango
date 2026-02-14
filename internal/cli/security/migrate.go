package security

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/langowarny/lango/internal/cli/prompt"
	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/logging"
	"github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
	"github.com/spf13/cobra"
)

var logger = logging.SubsystemSugar("cli-security")

// NewSecurityCmd creates the security command with lazy config loading.
func NewSecurityCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security",
		Short: "Manage security settings",
	}

	cmd.AddCommand(newMigratePassphraseCmd(cfgLoader))
	cmd.AddCommand(newSecretsCmd(cfgLoader))
	cmd.AddCommand(newStatusCmd(cfgLoader))

	return cmd
}

func newMigratePassphraseCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate-passphrase",
		Short: "Migrate encrypted data to a new passphrase",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if cfg.Security.Signer.Provider != "local" {
				return fmt.Errorf("this command is only available when using 'local' security provider")
			}

			if !prompt.IsInteractive() {
				return fmt.Errorf("this command requires an interactive terminal")
			}

			fmt.Println("This process will re-encrypt all your stored secrets with a new passphrase.")
			fmt.Println("Warning: If this process is interrupted, your data may be corrupted.")
			fmt.Println("Ensure you have a backup of your data directory.")
			fmt.Println()

			// 1. Initialize Store
			store, err := session.NewEntStore(cfg.Session.DatabasePath)
			if err != nil {
				return fmt.Errorf("failed to open session store: %w", err)
			}
			defer store.Close()

			// 2. Get Salt & Checksum
			salt, err := store.GetSalt("default")
			if err != nil {
				return fmt.Errorf("failed to get salt (is security initialized?): %w", err)
			}

			currentChecksum, err := store.GetChecksum("default")
			if err != nil {
				// Warn but proceed? Migration implies existing data.
				logger.Warn("no checksum found, proceeding without initial verification")
			}

			// 3. Prompt for Current Passphrase
			currentPass, err := prompt.Passphrase("Enter CURRENT passphrase: ")
			if err != nil {
				return err
			}

			// Verify Checksum
			provider := security.NewLocalCryptoProvider()
			if currentChecksum != nil {
				newChecksum := provider.CalculateChecksum(currentPass, salt)
				if !hmac.Equal(currentChecksum, newChecksum) {
					return fmt.Errorf("incorrect passphrase")
				}
			}

			// Initialize provider with old key
			if err := provider.InitializeWithSalt(currentPass, salt); err != nil {
				return fmt.Errorf("failed to initialize crypto provider: %w", err)
			}

			// 4. Prompt for New Passphrase
			newPass, err := prompt.PassphraseConfirm("Enter NEW passphrase: ", "Confirm NEW passphrase: ")
			if err != nil {
				return err
			}

			// 5. Perform Migration
			// Call migrateSecrets with Store
			ctx := context.Background()
			return migrateSecrets(ctx, store, currentPass, newPass, salt)
		},
	}
}

func migrateSecrets(ctx context.Context, store *session.EntStore, currentPass, newPass string, currentSalt []byte) error {
	// 1. Initialize Old Provider
	oldProvider := security.NewLocalCryptoProvider()
	if err := oldProvider.InitializeWithSalt(currentPass, currentSalt); err != nil {
		return fmt.Errorf("failed to init old provider: %w", err)
	}

	// 2. Generate new random salt
	newSalt := make([]byte, security.SaltSize)
	if _, err := io.ReadFull(rand.Reader, newSalt); err != nil {
		return fmt.Errorf("failed to generate new salt: %w", err)
	}

	// 3. Initialize New Provider with NEW salt
	newProvider := security.NewLocalCryptoProvider()
	if err := newProvider.InitializeWithSalt(newPass, newSalt); err != nil {
		return fmt.Errorf("failed to init new provider: %w", err)
	}

	// 4. Define Re-encryption Callback
	reencryptFn := func(ciphertext []byte) ([]byte, error) {
		plain, err := oldProvider.Decrypt(ctx, "local", ciphertext)
		if err != nil {
			return nil, err
		}
		return newProvider.Encrypt(ctx, "local", plain)
	}

	// 5. Calculate Checksum
	newChecksum := newProvider.CalculateChecksum(newPass, newSalt)

	// 6. Execute Migration in Store
	fmt.Println("Migrating secrets...")
	if err := store.MigrateSecrets(ctx, reencryptFn, newSalt, newChecksum); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migration completed successfully!")
	return nil
}
