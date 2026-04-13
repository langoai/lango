## ADDED Requirements

### Requirement: Non-interactive passphrase acquisition

The system SHALL provide a `passphrase.AcquireNonInteractive(opts Options)` function that acquires a passphrase only from keyring (Touch ID / TPM) or keyfile, without triggering any interactive terminal prompt or stdin pipe read. This function is used by commands that must work in non-interactive environments (e.g., `lango security status` default path).

#### Scenario: Acquire from keyring succeeds

- **WHEN** `AcquireNonInteractive` is called with a non-nil `KeyringProvider` that has a stored passphrase
- **THEN** the function returns the passphrase and `SourceKeyring`
- **AND** no interactive prompt is shown (biometric OS-level prompt is permitted)

#### Scenario: Fallback to keyfile when keyring has no value

- **WHEN** `AcquireNonInteractive` is called and the keyring returns `ErrNotFound`
- **AND** a keyfile exists at the configured path with 0600 permissions
- **THEN** the function returns the passphrase read from the keyfile and `SourceKeyfile`

#### Scenario: Error when neither source available

- **WHEN** `AcquireNonInteractive` is called and neither keyring nor keyfile provides a passphrase
- **THEN** the function returns an error without prompting for interactive input
- **AND** the caller is expected to gracefully degrade (e.g., show zero counts in status)

#### Scenario: Never reads stdin or tty

- **WHEN** `AcquireNonInteractive` is called in any environment
- **THEN** it SHALL NOT call `term.ReadPassword` (interactive prompt)
- **AND** it SHALL NOT read from `os.Stdin` pipe

### Requirement: Recovery credential choice during bootstrap

The system SHALL offer a credential choice when bootstrap loads an envelope that contains at least one mnemonic slot, in an interactive terminal. The user SHALL be able to select "passphrase" (default) or "recovery mnemonic" to unlock the envelope.

#### Scenario: Interactive bootstrap with mnemonic slot prompts choice

- **WHEN** bootstrap Phase 4 runs in an interactive terminal
- **AND** the loaded envelope contains a slot of type `KEKSlotMnemonic`
- **THEN** the user is prompted to choose between passphrase and mnemonic input
- **AND** choosing "mnemonic" skips `passphrase.Acquire()` entirely and sets `RecoveryMode = true`

#### Scenario: Non-interactive bootstrap skips choice prompt

- **WHEN** bootstrap runs non-interactively (no tty, keyring/keyfile available)
- **THEN** no choice prompt is shown
- **AND** passphrase acquisition proceeds via normal priority chain
