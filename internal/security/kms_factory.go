package security

import (
	"fmt"

	"github.com/langoai/lango/internal/config"
)

// NewKMSProvider creates a CryptoProvider for the named KMS backend.
// Supported providers: "aws-kms", "gcp-kms", "azure-kv", "pkcs11".
// Build tags control which providers are compiled in; uncompiled providers
// return a descriptive error.
func NewKMSProvider(providerName string, kmsConfig config.KMSConfig) (CryptoProvider, error) {
	switch providerName {
	case "aws-kms":
		return newAWSKMSProvider(kmsConfig)
	case "gcp-kms":
		return newGCPKMSProvider(kmsConfig)
	case "azure-kv":
		return newAzureKVProvider(kmsConfig)
	case "pkcs11":
		return newPKCS11Provider(kmsConfig)
	default:
		return nil, fmt.Errorf("unknown KMS provider: %q (supported: aws-kms, gcp-kms, azure-kv, pkcs11)", providerName)
	}
}
