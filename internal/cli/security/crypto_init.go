package security

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/bootstrap"
	sec "github.com/langoai/lango/internal/security"
)

// secretsStoreFromBoot creates a SecretsStore from a bootstrap result,
// using the already-initialized crypto provider. This avoids a second
// passphrase prompt since bootstrap already acquired and verified it.
func secretsStoreFromBoot(boot *bootstrap.Result) (*sec.SecretsStore, error) {
	ctx := context.Background()
	registry := sec.NewKeyRegistry(boot.DBClient)
	if _, err := registry.RegisterKey(ctx, "default", "local", sec.KeyTypeEncryption); err != nil {
		return nil, fmt.Errorf("register default key: %w", err)
	}
	return sec.NewSecretsStore(boot.DBClient, registry, boot.Crypto), nil
}
