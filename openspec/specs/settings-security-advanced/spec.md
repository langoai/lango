## Purpose

Define the TUI settings forms for advanced security configuration: OS keyring, SQLCipher DB encryption, and Cloud KMS / HSM backends.

## Requirements

### Requirement: Security Keyring settings form
The settings TUI SHALL provide a "Security Keyring" menu category with a single field for OS keyring enabled/disabled.

#### Scenario: User enables keyring
- **WHEN** user checks "OS Keyring Enabled"
- **THEN** the config's `security.keyring.enabled` SHALL be set to true

### Requirement: Security DB Encryption settings form
The settings TUI SHALL provide a read-only "Security DB Encryption" menu category that surfaces deprecated SQLCipher compatibility values without letting the operator treat them as active runtime controls.

#### Scenario: Deprecated DB encryption values are rendered read-only
- **WHEN** the Security DB Encryption category is displayed
- **THEN** the form SHALL render the legacy SQLCipher flag and cipher page size as non-editable informational fields
- **AND** the form SHALL explain that SQLCipher page encryption is unsupported by the current runtime
- **AND** save/update handling SHALL ignore both the current read-only status fields and the former editable DB encryption keys

### Requirement: Security KMS settings form
The settings TUI SHALL provide a "Security KMS" menu category with fields for region, key ID, endpoint, fallback to local, timeout, max retries, Azure vault URL, Azure key version, PKCS#11 module path, slot ID, PIN (password field), and key label.

#### Scenario: User configures AWS KMS
- **WHEN** user enters region "us-east-1" and a key ARN
- **THEN** the config's `security.kms.region` and `security.kms.keyId` SHALL contain the entered values

#### Scenario: PKCS#11 PIN is password field
- **WHEN** the KMS form is displayed
- **THEN** the PKCS#11 PIN field SHALL use InputPassword type to mask the value

#### Scenario: KMS fallback field explains bootstrap override
- **WHEN** the KMS form is displayed
- **THEN** the fallback-to-local field SHALL explain that profile-backed fallback covers KMS signing, encryption, and decryption after profile config is loaded
- **AND** it SHALL mention `LANGO_KMS_FALLBACK_TO_LOCAL=false` for fail-closed encrypted profile bootstrap KMS unwrap before profile config is loaded
