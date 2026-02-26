## ADDED Requirements

### Requirement: Security Keyring settings form
The settings TUI SHALL provide a "Security Keyring" menu category with a single field for OS keyring enabled/disabled.

#### Scenario: User enables keyring
- **WHEN** user checks "OS Keyring Enabled"
- **THEN** the config's `security.keyring.enabled` SHALL be set to true

### Requirement: Security DB Encryption settings form
The settings TUI SHALL provide a "Security DB Encryption" menu category with fields for SQLCipher encryption enabled and cipher page size.

#### Scenario: User enables DB encryption
- **WHEN** user checks "SQLCipher Encryption" and sets page size to 4096
- **THEN** the config's `security.dbEncryption.enabled` SHALL be true and `cipherPageSize` SHALL be 4096

#### Scenario: Cipher page size validation
- **WHEN** user enters 0 or a negative number for cipher page size
- **THEN** the form SHALL display a validation error "must be a positive integer"

### Requirement: Security KMS settings form
The settings TUI SHALL provide a "Security KMS" menu category with fields for region, key ID, endpoint, fallback to local, timeout, max retries, Azure vault URL, Azure key version, PKCS#11 module path, slot ID, PIN (password field), and key label.

#### Scenario: User configures AWS KMS
- **WHEN** user enters region "us-east-1" and a key ARN
- **THEN** the config's `security.kms.region` and `security.kms.keyId` SHALL contain the entered values

#### Scenario: PKCS#11 PIN is password field
- **WHEN** the KMS form is displayed
- **THEN** the PKCS#11 PIN field SHALL use InputPassword type to mask the value
