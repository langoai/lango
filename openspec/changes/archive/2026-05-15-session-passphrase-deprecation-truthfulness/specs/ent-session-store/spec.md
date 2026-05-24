## ADDED Requirements

### Requirement: Deprecated SQLCipher-style session passphrase input stays inert
The ent-backed session store SHALL NOT silently reactivate SQLCipher page-encryption behavior through the legacy passphrase constructor option now that the runtime uses plaintext SQLite plus higher-level payload protection.

#### Scenario: Plaintext session store opens with deprecated passphrase option
- **WHEN** `NewEntStore` is called with `WithPassphrase(...)` against a plaintext SQLite database
- **THEN** the store still opens successfully
- **AND** the deprecated passphrase option is treated as compatibility-only input rather than an active SQLCipher unlock path
