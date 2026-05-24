## Purpose

Capability spec for db-encryption. This capability now describes legacy encrypted-database handling after SQLCipher runtime support removal.

## Requirements

### Requirement: Legacy encrypted DB fail-fast
The runtime MUST reject non-SQLite database headers as `legacy encrypted or unreadable DB` and MUST surface a remediation-oriented error instead of attempting SQLCipher unlock.

#### Scenario: Legacy encrypted header detected during open
- **WHEN** the database file exists, is at least 16 bytes long, and does not begin with `SQLite format 3`
- **THEN** the runtime returns an error indicating `legacy encrypted or unreadable DB`
- **AND** the error message includes remediation guidance to downgrade or export the database

### Requirement: SQLCipher migration workflows removal
The runtime MUST eventually remove SQLCipher-specific encryption and decryption workflows from the supported database lifecycle.

#### Scenario: SQLCipher workflow disabled
- **WHEN** a SQLCipher-specific database migration or decrypt workflow is invoked after this change is complete
- **THEN** the command fails with a clear unsupported/remediation message

### Requirement: Deprecated SQLCipher open arguments remain compatibility-only
The runtime MUST NOT silently reactivate SQLCipher page-encryption behavior through legacy open-function arguments after SQLCipher runtime support removal.

#### Scenario: Plaintext open succeeds with deprecated encryption args
- **WHEN** a plaintext SQLite database is opened through the managed or read-only DB-open paths with legacy encryption arguments such as a passphrase, raw-key flag, or cipher page size
- **THEN** the runtime still opens the plaintext database successfully
- **AND** it does not enable SQLCipher page-level encryption behavior
