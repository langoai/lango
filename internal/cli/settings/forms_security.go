package settings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/types"
)

// NewSecurityForm creates the Security configuration form.
func NewSecurityForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Security Configuration")

	interceptorEnabled := &tuicore.Field{
		Key: "interceptor_enabled", Label: "Privacy Interceptor", Type: tuicore.InputBool,
		Checked:     cfg.Security.Interceptor.Enabled,
		Description: "Enable the privacy interceptor to filter outgoing data",
	}
	form.AddField(interceptorEnabled)
	isInterceptorOn := func() bool { return interceptorEnabled.Checked }

	form.AddField(&tuicore.Field{
		Key: "interceptor_pii", Label: "  Redact PII", Type: tuicore.InputBool,
		Checked:     cfg.Security.Interceptor.RedactPII,
		Description: "Automatically redact personally identifiable information from messages",
		VisibleWhen: isInterceptorOn,
	})
	policyVal := string(cfg.Security.Interceptor.ApprovalPolicy)
	if policyVal == "" {
		policyVal = "dangerous"
	}
	form.AddField(&tuicore.Field{
		Key: "interceptor_policy", Label: "  Approval Policy", Type: tuicore.InputSelect,
		Value:       policyVal,
		Options:     []string{"dangerous", "all", "configured", "none"},
		Description: "When to require user approval: dangerous=risky tools, all=every tool, none=skip",
		VisibleWhen: isInterceptorOn,
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_timeout", Label: "  Approval Timeout (s)", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Security.Interceptor.ApprovalTimeoutSec),
		Description: "Seconds to wait for user approval before auto-denying; 0 = wait forever",
		VisibleWhen: isInterceptorOn,
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_notify", Label: "  Notify Channel", Type: tuicore.InputSelect,
		Value:       cfg.Security.Interceptor.NotifyChannel,
		Options:     []string{"", string(types.ChannelTelegram), string(types.ChannelDiscord), string(types.ChannelSlack)},
		Description: "Channel to send approval notifications to; empty = no notification",
		VisibleWhen: isInterceptorOn,
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_sensitive_tools", Label: "  Sensitive Tools", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Security.Interceptor.SensitiveTools, ","),
		Placeholder: "exec,browser (comma-separated)",
		Description: "Tools that always require approval regardless of approval policy",
		VisibleWhen: isInterceptorOn,
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_exempt_tools", Label: "  Exempt Tools", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Security.Interceptor.ExemptTools, ","),
		Placeholder: "filesystem (comma-separated)",
		Description: "Tools that never require approval, even with 'all' policy",
		VisibleWhen: isInterceptorOn,
	})

	// PII Pattern Management
	form.AddField(&tuicore.Field{
		Key: "interceptor_pii_disabled", Label: "  Disabled PII Patterns", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Security.Interceptor.PIIDisabledPatterns, ","),
		Placeholder: "kr_bank_account,passport,ipv4 (comma-separated)",
		Description: "Built-in PII pattern names to disable (e.g. ipv4, passport)",
		VisibleWhen: isInterceptorOn,
	})
	form.AddField(&tuicore.Field{
		Key: "interceptor_pii_custom", Label: "  Custom PII Patterns", Type: tuicore.InputText,
		Value:       formatCustomPatterns(cfg.Security.Interceptor.PIICustomPatterns),
		Placeholder: `my_id:\bID-\d{6}\b (name:regex, comma-sep)`,
		Description: "Custom regex patterns for PII detection in name:regex format",
		VisibleWhen: isInterceptorOn,
	})

	// Presidio Integration
	presidioEnabled := &tuicore.Field{
		Key: "presidio_enabled", Label: "  Presidio (Docker)", Type: tuicore.InputBool,
		Checked:     cfg.Security.Interceptor.Presidio.Enabled,
		Description: "Use Microsoft Presidio (Docker) for advanced NLP-based PII detection",
		VisibleWhen: isInterceptorOn,
	}
	form.AddField(presidioEnabled)
	isPresidioOn := func() bool { return isInterceptorOn() && presidioEnabled.Checked }
	form.AddField(&tuicore.Field{
		Key: "presidio_url", Label: "    Presidio URL", Type: tuicore.InputText,
		Value:       cfg.Security.Interceptor.Presidio.URL,
		Placeholder: "http://localhost:5002",
		Description: "URL of the Presidio analyzer service endpoint",
		VisibleWhen: isPresidioOn,
	})
	presidioLang := cfg.Security.Interceptor.Presidio.Language
	if presidioLang == "" {
		presidioLang = "en"
	}
	form.AddField(&tuicore.Field{
		Key: "presidio_language", Label: "    Presidio Language", Type: tuicore.InputSelect,
		Value:       presidioLang,
		Options:     []string{"en", "ko", "ja", "zh", "de", "fr", "es", "it", "pt", "nl", "ru"},
		Description: "Primary language for Presidio NLP analysis",
		VisibleWhen: isPresidioOn,
	})

	// Signer Configuration
	signerField := &tuicore.Field{
		Key: "signer_provider", Label: "Signer Provider", Type: tuicore.InputSelect,
		Value:       cfg.Security.Signer.Provider,
		Options:     []string{"local", "rpc", "enclave", "aws-kms", "gcp-kms", "azure-kv", "pkcs11"},
		Description: "Cryptographic signer backend for message signing and verification",
	}
	form.AddField(signerField)
	form.AddField(&tuicore.Field{
		Key: "signer_rpc", Label: "  RPC URL", Type: tuicore.InputText,
		Value:       cfg.Security.Signer.RPCUrl,
		Placeholder: "http://localhost:8080",
		Description: "URL of the remote signing service",
		VisibleWhen: func() bool { return signerField.Value == "rpc" },
	})
	form.AddField(&tuicore.Field{
		Key: "signer_keyid", Label: "  Key ID", Type: tuicore.InputText,
		Value:       cfg.Security.Signer.KeyID,
		Placeholder: "key-123",
		Description: "Key identifier for the signer (ARN for AWS, key name for GCP/Azure)",
		VisibleWhen: func() bool {
			v := signerField.Value
			return v == "rpc" || v == "aws-kms" || v == "gcp-kms" || v == "azure-kv" || v == "pkcs11"
		},
	})

	return &form
}

// NewDBEncryptionForm creates the Security DB Encryption configuration form.
func NewDBEncryptionForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Legacy DB Encryption Configuration")

	form.AddField(&tuicore.Field{
		Key: "db_encryption_enabled", Label: "Legacy SQLCipher Flag", Type: tuicore.InputBool,
		Checked:     cfg.Security.DBEncryption.Enabled,
		Description: "Deprecated compatibility flag. The current runtime ignores SQLCipher database encryption settings.",
	})

	form.AddField(&tuicore.Field{
		Key: "db_cipher_page_size", Label: "Cipher Page Size", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Security.DBEncryption.CipherPageSize),
		Placeholder: "4096",
		Description: "Deprecated legacy SQLCipher page size. Retained only for parsing older configs.",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// NewKMSForm creates the Security KMS configuration form.
func NewKMSForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Security KMS Configuration")

	// Backend selector mirrors signer provider to drive field visibility.
	signerProv := cfg.Security.Signer.Provider
	if signerProv == "" {
		signerProv = "local"
	}
	backendField := &tuicore.Field{
		Key: "kms_backend", Label: "KMS Backend", Type: tuicore.InputSelect,
		Value:       signerProv,
		Options:     []string{"local", "aws-kms", "gcp-kms", "azure-kv", "pkcs11"},
		Description: "Cloud KMS or HSM backend; must match Signer Provider in Security settings",
	}
	form.AddField(backendField)

	isCloudKMS := func() bool {
		v := backendField.Value
		return v == "aws-kms" || v == "gcp-kms" || v == "azure-kv"
	}
	isAnyKMS := func() bool {
		return backendField.Value != "local"
	}

	form.AddField(&tuicore.Field{
		Key: "kms_region", Label: "Region", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Region,
		Placeholder: "us-east-1 or us-central1",
		Description: "Cloud region for KMS API calls (AWS region or GCP location)",
		VisibleWhen: isCloudKMS,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_key_id", Label: "Key ID", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.KeyID,
		Placeholder: "arn:aws:kms:... or alias/my-key",
		Description: "KMS key identifier (AWS ARN, GCP resource name, or Azure key name)",
		VisibleWhen: isAnyKMS,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_endpoint", Label: "Endpoint", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Endpoint,
		Placeholder: "http://localhost:8080 (optional)",
		Description: "Custom KMS API endpoint; leave empty for default cloud endpoints",
		VisibleWhen: isCloudKMS,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_fallback_to_local", Label: "Fallback to Local", Type: tuicore.InputBool,
		Checked:     cfg.Security.KMS.FallbackToLocal,
		Description: "Fall back to local key signing if cloud KMS is unavailable",
		VisibleWhen: isAnyKMS,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_timeout", Label: "Timeout Per Operation", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.TimeoutPerOperation.String(),
		Placeholder: "5s",
		Description: "Timeout for each individual KMS API call",
		VisibleWhen: isAnyKMS,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_max_retries", Label: "Max Retries", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Security.KMS.MaxRetries),
		Placeholder: "3",
		Description: "Number of retry attempts for failed KMS operations",
		VisibleWhen: isAnyKMS,
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	isAzure := func() bool { return backendField.Value == "azure-kv" }
	form.AddField(&tuicore.Field{
		Key: "kms_azure_vault_url", Label: "Azure Vault URL", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Azure.VaultURL,
		Placeholder: "https://myvault.vault.azure.net",
		Description: "Azure Key Vault URL (required for Azure backend)",
		VisibleWhen: isAzure,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_azure_key_version", Label: "Azure Key Version", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Azure.KeyVersion,
		Placeholder: "empty = latest",
		Description: "Specific key version to use; empty = always use latest version",
		VisibleWhen: isAzure,
	})

	isPKCS11 := func() bool { return backendField.Value == "pkcs11" }
	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_module", Label: "PKCS#11 Module Path", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.PKCS11.ModulePath,
		Placeholder: "/usr/lib/pkcs11/opensc-pkcs11.so",
		Description: "Path to the PKCS#11 shared library for HSM access",
		VisibleWhen: isPKCS11,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_slot_id", Label: "PKCS#11 Slot ID", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Security.KMS.PKCS11.SlotID),
		Placeholder: "0",
		Description: "HSM slot index to use for key operations",
		VisibleWhen: isPKCS11,
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_pin", Label: "PKCS#11 PIN", Type: tuicore.InputPassword,
		Value:       cfg.Security.KMS.PKCS11.Pin,
		Placeholder: "prefer LANGO_PKCS11_PIN env var",
		Description: "HSM PIN/password; strongly prefer LANGO_PKCS11_PIN env var for security",
		VisibleWhen: isPKCS11,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_key_label", Label: "PKCS#11 Key Label", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.PKCS11.KeyLabel,
		Placeholder: "my-signing-key",
		Description: "Label of the signing key stored in the HSM",
		VisibleWhen: isPKCS11,
	})

	return &form
}
