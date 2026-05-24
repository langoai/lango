## ADDED Requirements

### Requirement: Deprecated SQLCipher open arguments remain compatibility-only
The runtime MUST NOT silently reactivate SQLCipher page-encryption behavior through legacy open-function arguments after SQLCipher runtime support removal.

#### Scenario: Plaintext open succeeds with deprecated encryption args
- **WHEN** a plaintext SQLite database is opened through the managed or read-only DB-open paths with legacy encryption arguments such as a passphrase, raw-key flag, or cipher page size
- **THEN** the runtime still opens the plaintext database successfully
- **AND** it does not enable SQLCipher page-level encryption behavior
