## MODIFIED Requirements

### Requirement: Non-interactive passphrase acquisition

The system SHALL provide a `passphrase.AcquireNonInteractive(opts Options)` function that acquires a passphrase only from keyring (Touch ID / TPM) or keyfile, without triggering any interactive terminal prompt or stdin pipe read. This function is used by commands that must work in non-interactive environments (e.g., `lango security status` default path).

#### Scenario: Non-interactive keyring warning path supports injected writer
- **WHEN** the keyring read branch of `AcquireNonInteractive()` returns a non-`ErrNotFound` error in tests
- **THEN** the warning path SHALL be capturable via an injected writer instead of requiring `os.Stderr` interception

#### Scenario: Non-interactive ErrNotFound path stays silent
- **WHEN** the keyring read branch of `AcquireNonInteractive()` returns `ErrNotFound`
- **THEN** no warning SHALL be emitted and fallback to the next source SHALL continue
