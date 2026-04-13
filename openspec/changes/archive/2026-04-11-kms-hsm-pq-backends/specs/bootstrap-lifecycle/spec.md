## MODIFIED Requirements

### Requirement: Unified bootstrap sequence

#### Scenario: KMS-based passphraseless bootstrap
- **WHEN** an envelope has a `KEKSlotHardware` slot and KMS env vars (`LANGO_KMS_PROVIDER`, etc.) are set
- **THEN** the bootstrap SHALL create a bare KMS provider, attempt `UnwrapFromKMS` with 2-tier matching, and skip passphrase acquisition on success
- **AND** `Result.KMSUnwrap` SHALL be set to true

#### Scenario: KMS bootstrap with graceful fallback
- **WHEN** KMS unwrap fails (provider init error, decrypt error, or no matching slot)
- **THEN** the bootstrap SHALL print a warning to stderr
- **AND** fall through to the standard credential acquisition (mnemonic → passphrase)
- **AND** NOT attempt local decryption of KMS-wrapped ciphertext

#### Scenario: No KMS env vars configured
- **WHEN** `LANGO_KMS_PROVIDER` is not set and `Options.KMSConfig` is nil
- **THEN** the bootstrap SHALL follow the standard passphrase path without any KMS attempt

#### Scenario: Explicit Options override env vars
- **WHEN** `Options.KMSConfig` and `Options.KMSProviderName` are explicitly set
- **THEN** the bootstrap SHALL use those values and NOT read KMS env vars
- **AND** env vars serve only as default source when Options are empty

## ADDED Requirements

### Requirement: KMS bootstrap env config

The system SHALL provide a `KMSConfigFromEnv()` function that reads KMS KEK configuration from environment variables. Provider-specific env vars:

#### Scenario: AWS KMS / GCP KMS env config
- **WHEN** `LANGO_KMS_PROVIDER` is `aws-kms` or `gcp-kms`
- **THEN** the function reads `LANGO_KMS_KEY_ID`, `LANGO_KMS_REGION`, `LANGO_KMS_ENDPOINT`

#### Scenario: Azure Key Vault env config
- **WHEN** `LANGO_KMS_PROVIDER` is `azure-kv`
- **THEN** the function reads `LANGO_KMS_KEY_ID`, `LANGO_KMS_AZURE_VAULT_URL`, `LANGO_KMS_AZURE_KEY_VERSION`

#### Scenario: PKCS#11 HSM env config
- **WHEN** `LANGO_KMS_PROVIDER` is `pkcs11`
- **THEN** the function reads `LANGO_KMS_PKCS11_MODULE`, `LANGO_KMS_PKCS11_SLOT_ID`, `LANGO_KMS_PKCS11_KEY_LABEL`, `LANGO_PKCS11_PIN`

#### Scenario: Missing provider env var
- **WHEN** `LANGO_KMS_PROVIDER` is not set
- **THEN** the function returns nil config and empty provider name
