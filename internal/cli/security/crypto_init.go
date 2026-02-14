package security

import (
	"context"
	"crypto/hmac"
	"fmt"

	"github.com/langowarny/lango/internal/passphrase"
	sec "github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
)

// initLocalCrypto creates a SecretsStore with local crypto, handling passphrase
// resolution, salt loading/generation, and checksum verification.
// Passphrase is acquired via the passphrase package (keyfile → interactive → stdin).
func initLocalCrypto(store *session.EntStore) (*sec.SecretsStore, error) {
	// 1. Acquire passphrase via priority chain
	pass, _, err := passphrase.Acquire(passphrase.Options{})
	if err != nil {
		return nil, fmt.Errorf("acquire passphrase: %w", err)
	}

	provider := sec.NewLocalCryptoProvider()

	// 2. Load or generate salt
	salt, err := store.GetSalt("default")
	if err != nil {
		// First-time setup: initialize with new salt
		if err := provider.Initialize(pass); err != nil {
			return nil, fmt.Errorf("initialize crypto: %w", err)
		}
		if err := store.SetSalt("default", provider.Salt()); err != nil {
			return nil, fmt.Errorf("store salt: %w", err)
		}
		checksum := provider.CalculateChecksum(pass, provider.Salt())
		if err := store.SetChecksum("default", checksum); err != nil {
			return nil, fmt.Errorf("store checksum: %w", err)
		}
	} else {
		// 3. Verify checksum
		storedChecksum, csErr := store.GetChecksum("default")
		if csErr == nil && storedChecksum != nil {
			newChecksum := provider.CalculateChecksum(pass, salt)
			if !hmac.Equal(storedChecksum, newChecksum) {
				return nil, fmt.Errorf("incorrect passphrase")
			}
		}

		if err := provider.InitializeWithSalt(pass, salt); err != nil {
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
