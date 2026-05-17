## MODIFIED Requirements

### Requirement: Passphrase acquisition stream seams

The stdin-pipe and keyring-warning branches of passphrase acquisition SHALL be structured so they can be exercised with injected readers and writers in tests, without replacing process-global stdin or stderr. Public acquisition wrappers SHALL route their process stdio dependencies through package-level seams before delegating to lower-level helpers.

#### Scenario: Public acquire wrapper supports injected stdin seam

- **WHEN** `Acquire(...)` falls through to the stdin-pipe branch with terminal detection disabled
- **THEN** it SHALL read the passphrase from the injected stdin seam
- **AND** it SHALL NOT require replacing process-global `os.Stdin`

#### Scenario: Public acquire wrapper supports injected stderr seam

- **WHEN** `Acquire(...)` observes a keyring read error other than `ErrNotFound`
- **THEN** it SHALL write the warning to the injected stderr seam
- **AND** it SHALL NOT require intercepting process-global `os.Stderr`

#### Scenario: Stdin pipe path supports injected reader

- **WHEN** the stdin-pipe branch of acquisition is exercised in tests
- **THEN** the implementation SHALL be able to read from an injected reader instead of requiring `os.Stdin` replacement

#### Scenario: Keyring warning path supports injected writer

- **WHEN** the keyring read branch returns a non-`ErrNotFound` error in tests
- **THEN** the warning path SHALL be capturable via an injected writer instead of requiring `os.Stderr` interception

#### Scenario: Keyring returns ErrNotFound

- **WHEN** `KeyringProvider.Get()` returns `ErrNotFound`
- **THEN** no warning SHALL be written to stderr
- **AND** acquisition SHALL continue to the next source

### Requirement: Non-interactive passphrase acquisition

The system SHALL provide a `passphrase.AcquireNonInteractive(opts Options)` function that acquires a passphrase only from keyring (Touch ID / TPM) or keyfile, without triggering any interactive terminal prompt or stdin pipe read. This function is used by commands that must work in non-interactive environments (e.g., `lango security status` default path). Warning output emitted by the public wrapper SHALL use the package stderr seam.

#### Scenario: Public non-interactive wrapper supports injected stderr seam

- **WHEN** `AcquireNonInteractive(...)` observes a keyring read error other than `ErrNotFound`
- **THEN** it SHALL write the warning to the injected stderr seam
- **AND** it SHALL NOT require intercepting process-global `os.Stderr`
