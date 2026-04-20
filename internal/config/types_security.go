package config

import "time"

// SecurityConfig defines security settings
type SecurityConfig struct {
	// Interceptor configuration
	Interceptor InterceptorConfig `mapstructure:"interceptor" json:"interceptor"`
	// Exportability configuration
	Exportability ExportabilityConfig `mapstructure:"exportability" json:"exportability"`
	// Signer configuration
	Signer SignerConfig `mapstructure:"signer" json:"signer"`
	// DBEncryption configuration (SQLCipher transparent encryption)
	DBEncryption DBEncryptionConfig `mapstructure:"dbEncryption" json:"dbEncryption"`
	// KMS configuration (Cloud KMS / HSM backends)
	KMS KMSConfig `mapstructure:"kms" json:"kms"`
}

// ExportabilityConfig defines exportability policy settings.
type ExportabilityConfig struct {
	Enabled bool `mapstructure:"enabled" json:"enabled"`
}

// KMSConfig defines Cloud KMS and HSM backend settings.
type KMSConfig struct {
	// Region is the cloud region for KMS API calls (e.g. "us-east-1", "us-central1").
	Region string `mapstructure:"region" json:"region"`

	// KeyID is the KMS key identifier (ARN, resource name, or alias).
	KeyID string `mapstructure:"keyId" json:"keyId"`

	// Endpoint is an optional custom endpoint for KMS API calls (useful for testing).
	Endpoint string `mapstructure:"endpoint" json:"endpoint,omitempty"`

	// FallbackToLocal enables automatic fallback to the local CryptoProvider when KMS is unavailable.
	FallbackToLocal bool `mapstructure:"fallbackToLocal" json:"fallbackToLocal"`

	// TimeoutPerOperation is the maximum duration for a single KMS API call (default: 5s).
	TimeoutPerOperation time.Duration `mapstructure:"timeoutPerOperation" json:"timeoutPerOperation"`

	// MaxRetries is the number of retry attempts for transient KMS errors (default: 3).
	MaxRetries int `mapstructure:"maxRetries" json:"maxRetries"`

	// Azure holds Azure Key Vault specific settings.
	Azure AzureKVConfig `mapstructure:"azure" json:"azure"`

	// PKCS11 holds PKCS#11 HSM specific settings.
	PKCS11 PKCS11Config `mapstructure:"pkcs11" json:"pkcs11"`
}

// AzureKVConfig defines Azure Key Vault specific settings.
type AzureKVConfig struct {
	// VaultURL is the Azure Key Vault URL (e.g. "https://myvault.vault.azure.net").
	VaultURL string `mapstructure:"vaultUrl" json:"vaultUrl"`

	// KeyVersion is the specific key version to use (empty = latest).
	KeyVersion string `mapstructure:"keyVersion" json:"keyVersion,omitempty"`
}

// PKCS11Config defines PKCS#11 HSM specific settings.
type PKCS11Config struct {
	// ModulePath is the path to the PKCS#11 shared library (.so/.dylib/.dll).
	ModulePath string `mapstructure:"modulePath" json:"modulePath"`

	// SlotID is the PKCS#11 slot number to use.
	SlotID int `mapstructure:"slotId" json:"slotId"`

	// Pin is the PKCS#11 user PIN (prefer LANGO_PKCS11_PIN env var).
	Pin string `mapstructure:"pin" json:"pin,omitempty"`

	// KeyLabel is the label of the key object in the HSM.
	KeyLabel string `mapstructure:"keyLabel" json:"keyLabel"`
}

// DBEncryptionConfig defines SQLCipher transparent database encryption settings.
type DBEncryptionConfig struct {
	// Enabled activates SQLCipher encryption for the application database.
	Enabled bool `mapstructure:"enabled" json:"enabled"`
	// CipherPageSize is the SQLCipher cipher_page_size PRAGMA (default: 4096).
	CipherPageSize int `mapstructure:"cipherPageSize" json:"cipherPageSize"`
}

// ApprovalPolicy determines which tools require approval before execution.
type ApprovalPolicy string

const (
	// ApprovalPolicyDangerous requires approval for Dangerous-level tools (default).
	ApprovalPolicyDangerous ApprovalPolicy = "dangerous"
	// ApprovalPolicyAll requires approval for all tools.
	ApprovalPolicyAll ApprovalPolicy = "all"
	// ApprovalPolicyConfigured requires approval only for explicitly listed SensitiveTools.
	ApprovalPolicyConfigured ApprovalPolicy = "configured"
	// ApprovalPolicyNone disables approval entirely.
	ApprovalPolicyNone ApprovalPolicy = "none"
)

// Valid reports whether p is a known approval policy.
func (p ApprovalPolicy) Valid() bool {
	switch p {
	case ApprovalPolicyDangerous, ApprovalPolicyAll, ApprovalPolicyConfigured, ApprovalPolicyNone:
		return true
	}
	return false
}

// Values returns all known approval policies.
func (p ApprovalPolicy) Values() []ApprovalPolicy {
	return []ApprovalPolicy{ApprovalPolicyDangerous, ApprovalPolicyAll, ApprovalPolicyConfigured, ApprovalPolicyNone}
}

// InterceptorConfig defines AI Privacy Interceptor settings
type InterceptorConfig struct {
	Enabled             bool              `mapstructure:"enabled" json:"enabled"`
	RedactPII           bool              `mapstructure:"redactPii" json:"redactPii"`
	ApprovalPolicy      ApprovalPolicy    `mapstructure:"approvalPolicy" json:"approvalPolicy"` // default: "dangerous"
	HeadlessAutoApprove bool              `mapstructure:"headlessAutoApprove" json:"headlessAutoApprove"`
	NotifyChannel       string            `mapstructure:"notifyChannel" json:"notifyChannel"` // e.g. "discord", "telegram"
	SensitiveTools      []string          `mapstructure:"sensitiveTools" json:"sensitiveTools"`
	ExemptTools         []string          `mapstructure:"exemptTools" json:"exemptTools"` // Tools exempt from approval regardless of policy
	PIIRegexPatterns    []string          `mapstructure:"piiRegexPatterns" json:"piiRegexPatterns"`
	ApprovalTimeoutSec  int               `mapstructure:"approvalTimeoutSec" json:"approvalTimeoutSec"` // default 30
	PIIDisabledPatterns []string          `mapstructure:"piiDisabledPatterns" json:"piiDisabledPatterns"`
	PIICustomPatterns   map[string]string `mapstructure:"piiCustomPatterns" json:"piiCustomPatterns"`
	Presidio            PresidioConfig    `mapstructure:"presidio" json:"presidio"`
}

// PresidioConfig defines Microsoft Presidio integration settings.
type PresidioConfig struct {
	Enabled        bool    `mapstructure:"enabled" json:"enabled"`
	URL            string  `mapstructure:"url" json:"url"`                       // default: http://localhost:5002
	ScoreThreshold float64 `mapstructure:"scoreThreshold" json:"scoreThreshold"` // default: 0.7
	Language       string  `mapstructure:"language" json:"language"`             // default: "en"
}

// SignerConfig defines Secure Signer settings
type SignerConfig struct {
	Provider string `mapstructure:"provider" json:"provider"` // "local", "rpc", "enclave"
	RPCUrl   string `mapstructure:"rpcUrl" json:"rpcUrl"`     // for RPC provider
	KeyID    string `mapstructure:"keyId" json:"keyId"`       // Key identifier
}

// AuthConfig defines authentication settings
type AuthConfig struct {
	// OIDC Providers
	Providers map[string]OIDCProviderConfig `mapstructure:"providers" json:"providers"`
}

// OIDCProviderConfig defines a single OIDC provider
type OIDCProviderConfig struct {
	IssuerURL    string   `mapstructure:"issuerUrl" json:"issuerUrl"`
	ClientID     string   `mapstructure:"clientId" json:"clientId"`
	ClientSecret string   `mapstructure:"clientSecret" json:"clientSecret"`
	RedirectURL  string   `mapstructure:"redirectUrl" json:"redirectUrl"`
	Scopes       []string `mapstructure:"scopes" json:"scopes"`
}
