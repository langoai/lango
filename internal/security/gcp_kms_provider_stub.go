//go:build !kms_gcp && !kms_all

package security

import (
	"fmt"

	"github.com/langoai/lango/internal/config"
)

func newGCPKMSProvider(_ config.KMSConfig) (CryptoProvider, error) {
	return nil, fmt.Errorf("GCP KMS support not compiled: rebuild with -tags kms_gcp")
}
