package bootstrap

import (
	"os"
	"strconv"

	"github.com/langoai/lango/internal/config"
)

// KMS bootstrap environment variable names.
const (
	envKMSProvider        = "LANGO_KMS_PROVIDER"
	envKMSKeyID           = "LANGO_KMS_KEY_ID"
	envKMSRegion          = "LANGO_KMS_REGION"
	envKMSEndpoint        = "LANGO_KMS_ENDPOINT"
	envKMSAzureVaultURL   = "LANGO_KMS_AZURE_VAULT_URL"
	envKMSAzureKeyVersion = "LANGO_KMS_AZURE_KEY_VERSION"
	envKMSPKCS11Module    = "LANGO_KMS_PKCS11_MODULE"
	envKMSPKCS11SlotID    = "LANGO_KMS_PKCS11_SLOT_ID"
	envKMSPKCS11KeyLabel  = "LANGO_KMS_PKCS11_KEY_LABEL"
	envPKCS11PIN          = "LANGO_PKCS11_PIN" // existing env var, reused
)

// KMSConfigFromEnv reads KMS KEK configuration from environment variables.
// Returns nil config and empty provider name if LANGO_KMS_PROVIDER is not set.
//
// Provider-specific env vars:
//
//	aws-kms/gcp-kms: LANGO_KMS_KEY_ID, LANGO_KMS_REGION, LANGO_KMS_ENDPOINT
//	azure-kv:        + LANGO_KMS_AZURE_VAULT_URL, LANGO_KMS_AZURE_KEY_VERSION
//	pkcs11:          LANGO_KMS_PKCS11_MODULE, LANGO_KMS_PKCS11_SLOT_ID,
//	                 LANGO_KMS_PKCS11_KEY_LABEL, LANGO_PKCS11_PIN
func KMSConfigFromEnv() (*config.KMSConfig, string) {
	provider := os.Getenv(envKMSProvider)
	if provider == "" {
		return nil, ""
	}

	cfg := &config.KMSConfig{
		KeyID:    os.Getenv(envKMSKeyID),
		Region:   os.Getenv(envKMSRegion),
		Endpoint: os.Getenv(envKMSEndpoint),
	}

	switch provider {
	case "azure-kv":
		cfg.Azure.VaultURL = os.Getenv(envKMSAzureVaultURL)
		cfg.Azure.KeyVersion = os.Getenv(envKMSAzureKeyVersion)
	case "pkcs11":
		cfg.PKCS11.ModulePath = os.Getenv(envKMSPKCS11Module)
		if s := os.Getenv(envKMSPKCS11SlotID); s != "" {
			if v, err := strconv.Atoi(s); err == nil {
				cfg.PKCS11.SlotID = v
			}
		}
		cfg.PKCS11.KeyLabel = os.Getenv(envKMSPKCS11KeyLabel)
		cfg.PKCS11.Pin = os.Getenv(envPKCS11PIN)
	}

	return cfg, provider
}
