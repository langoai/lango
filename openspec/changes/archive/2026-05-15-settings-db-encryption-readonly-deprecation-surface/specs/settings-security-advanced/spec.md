## MODIFIED Requirements

### Requirement: Security DB Encryption settings form
The settings TUI SHALL provide a read-only "Security DB Encryption" menu category that surfaces deprecated SQLCipher compatibility values without letting the operator treat them as active runtime controls.

#### Scenario: Deprecated DB encryption values are rendered read-only
- **WHEN** the Security DB Encryption category is displayed
- **THEN** the form SHALL render the legacy SQLCipher flag and cipher page size as non-editable informational fields
- **AND** the form SHALL explain that SQLCipher page encryption is unsupported by the current runtime
- **AND** save/update handling SHALL ignore both the current read-only status fields and the former editable DB encryption keys
