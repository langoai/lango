package security

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
)

var logger = logging.SubsystemSugar("cli-security")

// NewSecurityCmd creates the security command with lazy bootstrap loading.
func NewSecurityCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security",
		Short: "Manage security settings",
	}

	cmd.AddCommand(newMigratePassphraseCmd(bootLoader))
	cmd.AddCommand(newSecretsCmd(bootLoader))
	cmd.AddCommand(newStatusCmd(bootLoader))
	cmd.AddCommand(newKeyringCmd(bootLoader))

	return cmd
}

func newMigratePassphraseCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate-passphrase",
		Short: "Migrate encrypted data to a new passphrase",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			if boot.Config.Security.Signer.Provider != "local" {
				return fmt.Errorf("this command is only available when using 'local' security provider")
			}

			if !prompt.IsInteractive() {
				return fmt.Errorf("this command requires an interactive terminal")
			}

			fmt.Println("This process will re-encrypt all your stored secrets with a new passphrase.")
			fmt.Println("Warning: If this process is interrupted, your data may be corrupted.")
			fmt.Println("Ensure you have a backup of your data directory.")
			fmt.Println()

			store := session.NewEntStoreWithClient(boot.DBClient)

			// Current passphrase already verified by bootstrap
			newPass, err := prompt.PassphraseConfirm("Enter NEW passphrase: ", "Confirm NEW passphrase: ")
			if err != nil {
				return err
			}

			ctx := context.Background()
			return migrateSecrets(ctx, store, boot.Crypto, newPass)
		},
	}
}

func migrateSecrets(ctx context.Context, store *session.EntStore, oldCrypto security.CryptoProvider, newPass string) error {
	// Generate new random salt
	newSalt := make([]byte, security.SaltSize)
	if _, err := io.ReadFull(rand.Reader, newSalt); err != nil {
		return fmt.Errorf("generate new salt: %w", err)
	}

	// Initialize new provider with new salt
	newProvider := security.NewLocalCryptoProvider()
	if err := newProvider.InitializeWithSalt(newPass, newSalt); err != nil {
		return fmt.Errorf("init new provider: %w", err)
	}

	// Define re-encryption callback
	reencryptFn := func(ciphertext []byte) ([]byte, error) {
		plain, err := oldCrypto.Decrypt(ctx, "local", ciphertext)
		if err != nil {
			return nil, err
		}
		return newProvider.Encrypt(ctx, "local", plain)
	}

	// Calculate new checksum
	newChecksum := newProvider.CalculateChecksum(newPass, newSalt)

	// Execute migration in store
	fmt.Println("Migrating secrets...")
	if err := store.MigrateSecrets(ctx, reencryptFn, newSalt, newChecksum); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migration completed successfully!")
	return nil
}
