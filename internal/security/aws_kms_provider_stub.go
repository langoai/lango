//go:build !kms_aws && !kms_all

package security

import (
	"fmt"

	"github.com/langoai/lango/internal/config"
)

func newAWSKMSProvider(_ config.KMSConfig) (CryptoProvider, error) {
	return nil, fmt.Errorf("AWS KMS support not compiled: rebuild with -tags kms_aws")
}
