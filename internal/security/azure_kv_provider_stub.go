//go:build !kms_azure && !kms_all

package security

import (
	"fmt"

	"github.com/langoai/lango/internal/config"
)

func newAzureKVProvider(_ config.KMSConfig) (CryptoProvider, error) {
	return nil, fmt.Errorf("Azure Key Vault support not compiled: rebuild with -tags kms_azure")
}
