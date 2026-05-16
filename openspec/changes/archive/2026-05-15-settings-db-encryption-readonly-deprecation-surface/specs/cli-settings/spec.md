## MODIFIED Requirements

### Requirement: Security DB Encryption settings form
The settings TUI SHALL provide a read-only "Security DB Encryption" notice form that shows deprecated legacy SQLCipher config values without presenting them as active editable runtime settings.

#### Scenario: Deprecated DB encryption settings are shown read-only
- **WHEN** user opens the "Security DB Encryption" form
- **THEN** the form SHALL show read-only status rows for the legacy SQLCipher flag and cipher page size
- **AND** it SHALL include an explicit runtime note that SQLCipher page encryption is unsupported and broker-managed payload protection is active
- **AND** saving or updating the form SHALL NOT mutate `security.dbEncryption.*` runtime config values
