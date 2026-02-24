## MODIFIED Requirements

### Requirement: Composite Provider Strategy
The system SHALL use a composite provider that tries the primary provider first, then falls back to local. The primary provider MAY be a companion (RPC), Cloud KMS, or PKCS#11 backend.

#### Scenario: Companion available
- **WHEN** companion is connected
- **THEN** the system SHALL delegate crypto operations to companion via RPCProvider

#### Scenario: Companion unavailable with fallback
- **WHEN** companion is not connected
- **AND** local fallback is configured
- **AND** terminal is interactive (TTY available)
- **THEN** the system SHALL use local provider

#### Scenario: KMS primary with local fallback
- **WHEN** a KMS provider is configured as `security.signer.provider`
- **AND** `security.kms.fallbackToLocal` is true
- **THEN** the system SHALL wrap KMS in CompositeCryptoProvider with local as fallback and KMSHealthChecker as ConnectionChecker

## ADDED Requirements

### Requirement: KMS Provider Configuration Validation
The config validator SHALL accept `aws-kms`, `gcp-kms`, `azure-kv`, and `pkcs11` as valid values for `security.signer.provider`. Provider-specific fields SHALL be validated when the corresponding provider is selected.

#### Scenario: AWS KMS requires keyId
- **WHEN** `security.signer.provider` is `aws-kms`
- **AND** `security.kms.keyId` is empty
- **THEN** config validation SHALL fail with a descriptive error

#### Scenario: Azure KV requires vaultUrl and keyId
- **WHEN** `security.signer.provider` is `azure-kv`
- **AND** `security.kms.azure.vaultUrl` is empty
- **THEN** config validation SHALL fail with a descriptive error

#### Scenario: PKCS#11 requires modulePath
- **WHEN** `security.signer.provider` is `pkcs11`
- **AND** `security.kms.pkcs11.modulePath` is empty
- **THEN** config validation SHALL fail with a descriptive error
