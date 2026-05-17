## ADDED Requirements

### Requirement: Public configuration docs include Cloud KMS settings
The public configuration references SHALL include the profile-backed `security.kms.*` settings and the bootstrap-time `LANGO_KMS_FALLBACK_TO_LOCAL=false` override.

#### Scenario: Configuration reference lists KMS settings
- **WHEN** a user reads `README.md` or `docs/configuration.md`
- **THEN** they find `security.kms.region`, `security.kms.keyId`, `security.kms.endpoint`, `security.kms.fallbackToLocal`, `security.kms.timeoutPerOperation`, `security.kms.maxRetries`, `security.kms.azure.vaultUrl`, `security.kms.azure.keyVersion`, `security.kms.pkcs11.modulePath`, `security.kms.pkcs11.slotId`, `security.kms.pkcs11.pin`, and `security.kms.pkcs11.keyLabel`

#### Scenario: Configuration reference explains bootstrap fallback override
- **WHEN** a user reads the Cloud KMS configuration section in `README.md` or `docs/configuration.md`
- **THEN** it explains that encrypted profile bootstrap must use `LANGO_KMS_FALLBACK_TO_LOCAL=false` with `LANGO_KMS_PROVIDER` to fail closed before profile config is loaded
