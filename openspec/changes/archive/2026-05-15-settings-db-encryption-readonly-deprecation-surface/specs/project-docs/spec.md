## MODIFIED Requirements

### Requirement: Security feature card stays current with active encryption model
The docs/index.md Security card SHALL mention hardware keyring, broker-managed payload protection, and Cloud KMS without presenting SQLCipher page encryption as an active current-runtime feature.

#### Scenario: docs/index.md Security card is complete
- **WHEN** a user reads the Security card on docs/index.md
- **THEN** it SHALL mention hardware keyring (Touch ID / TPM), broker-managed payload protection, and Cloud KMS integration
- **AND** it SHALL NOT present SQLCipher page encryption as supported by the current runtime

### Requirement: README Features security line stays current with active encryption model
The README.md Features section security line SHALL mention hardware keyring, broker-managed payload protection, and Cloud KMS without presenting SQLCipher page encryption as an active current-runtime feature.

#### Scenario: README security feature is complete
- **WHEN** a user reads the Features section of README.md
- **THEN** the Secure line SHALL include hardware keyring, broker-managed payload protection, and Cloud KMS
- **AND** any SQLCipher references SHALL be framed as legacy compatibility or remediation only
