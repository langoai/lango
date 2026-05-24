## ADDED Requirements

### Requirement: Deprecated session-passphrase behavior stays executable
Repository-level regressions that make the legacy session-store passphrase option behave like an active SQLCipher unlock path again SHALL be prevented by executable package tests.

#### Scenario: Plaintext session store ignores deprecated passphrase option
- **WHEN** the repository still ships `internal/session.NewEntStore` with `WithPassphrase(...)`
- **THEN** executable package tests SHALL fail if a plaintext session store starts failing solely because the deprecated passphrase option was provided
