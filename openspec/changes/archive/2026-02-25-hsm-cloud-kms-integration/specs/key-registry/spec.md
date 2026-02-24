## ADDED Requirements

### Requirement: KMS Key Registration in Wiring
When a KMS provider is initialized, the system SHALL register the KMS key in KeyRegistry with the cloud key ARN/ID as `RemoteKeyID` and name `kms-default`.

#### Scenario: KMS provider wiring registers key
- **WHEN** `initSecurity()` initializes a KMS provider (aws-kms, gcp-kms, azure-kv, pkcs11)
- **THEN** a key named `kms-default` SHALL be registered in KeyRegistry
- **AND** its RemoteKeyID SHALL be set to `security.kms.keyId`
- **AND** its type SHALL be `encryption`
