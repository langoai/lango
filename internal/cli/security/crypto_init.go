package security

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"fmt"
	"io"
	"os"

	"github.com/langowarny/lango/internal/cli/prompt"
	"github.com/langowarny/lango/internal/config"
	sec "github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
)

// initLocalCrypto creates a SecretsStore with local crypto, handling passphrase resolution,
// salt loading/generation, and checksum verification.
func initLocalCrypto(cfg *config.Config, store *session.EntStore) (*sec.SecretsStore, error) {
	// 1. Resolve passphrase: env > config > interactive prompt
	passphrase := os.Getenv("LANGO_PASSPHRASE")
	if passphrase == "" {
		passphrase = cfg.Security.Passphrase
	}
	if passphrase == "" {
		if !prompt.IsInteractive() {
			return nil, fmt.Errorf("no passphrase available (set LANGO_PASSPHRASE or use interactive terminal)")
		}
		var err error
		passphrase, err = prompt.Passphrase("Enter passphrase: ")
		if err != nil {
			return nil, fmt.Errorf("read passphrase: %w", err)
		}
	}

	provider := sec.NewLocalCryptoProvider()

	// 2. Load or generate salt
	salt, err := store.GetSalt("default")
	if err != nil {
		// First-time setup: generate salt
		salt = make([]byte, sec.SaltSize)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, fmt.Errorf("generate salt: %w", err)
		}
		if err := store.SetSalt("default", salt); err != nil {
			return nil, fmt.Errorf("store salt: %w", err)
		}

		// Initialize provider and store checksum
		if err := provider.InitializeWithSalt(passphrase, salt); err != nil {
			return nil, fmt.Errorf("initialize crypto: %w", err)
		}
		checksum := provider.CalculateChecksum(passphrase, salt)
		if err := store.SetChecksum("default", checksum); err != nil {
			return nil, fmt.Errorf("store checksum: %w", err)
		}
	} else {
		// 3. Verify checksum
		storedChecksum, err := store.GetChecksum("default")
		if err == nil && storedChecksum != nil {
			newChecksum := provider.CalculateChecksum(passphrase, salt)
			if !hmac.Equal(storedChecksum, newChecksum) {
				return nil, fmt.Errorf("incorrect passphrase")
			}
		}

		if err := provider.InitializeWithSalt(passphrase, salt); err != nil {
			return nil, fmt.Errorf("initialize crypto: %w", err)
		}
	}

	// 4. Create key registry and register default key
	ctx := context.Background()
	registry := sec.NewKeyRegistry(store.Client())
	if _, err := registry.RegisterKey(ctx, "default", "local", sec.KeyTypeEncryption); err != nil {
		return nil, fmt.Errorf("register default key: %w", err)
	}

	return sec.NewSecretsStore(store.Client(), registry, provider), nil
}
