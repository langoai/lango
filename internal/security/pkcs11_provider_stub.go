//go:build !kms_pkcs11 && !kms_all

package security

import (
	"fmt"

	"github.com/langoai/lango/internal/config"
)

func newPKCS11Provider(_ config.KMSConfig) (CryptoProvider, error) {
	return nil, fmt.Errorf("PKCS#11 support not compiled: rebuild with -tags kms_pkcs11")
}
