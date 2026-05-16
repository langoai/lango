# Recovery Mnemonic

## Purpose

BIP39 24-word recovery mnemonic support for the Master Key envelope. Allows users to regain access to all encrypted data if they lose their passphrase, by deriving a separate KEK from the mnemonic and storing it as an additional envelope slot.
## Requirements
### Requirement: Recovery mnemonic generation

The system SHALL generate 24-word BIP39 recovery mnemonics using the `github.com/tyler-smith/go-bip39` library. The mnemonic SHALL derive a KEK via the same `DeriveKEK(secret, slot)` dispatch used by passphrase slots; the slot's `Domain` field (e.g. `"mnemonic"`) provides cryptographic separation from passphrase KEKs, and each slot has its own unique PBKDF2 salt.

#### Scenario: Generated mnemonic is valid BIP39

- **WHEN** `GenerateRecoveryMnemonic()` is called
- **THEN** it returns a 24-word mnemonic string
- **AND** `ValidateMnemonic(mnemonic)` returns nil

#### Scenario: Each call produces a different mnemonic

- **WHEN** `GenerateRecoveryMnemonic()` is called twice
- **THEN** the two returned mnemonics differ with overwhelming probability

### Requirement: Mnemonic slot addition

The system SHALL provide `lango security recovery setup` command that generates a new mnemonic, adds it as a KEK slot to the envelope, displays the mnemonic to the user, and verifies user recording via confirmation word prompts.

#### Scenario: Setup adds mnemonic slot

- **WHEN** the user runs `lango security recovery setup`
- **AND** confirms the mnemonic by entering 2 randomly-requested confirmation words
- **THEN** a new KEK slot with `Type = KEKSlotMnemonic` is added to the envelope
- **AND** the envelope file is updated on disk

#### Scenario: Setup rejects incorrect confirmation words

- **WHEN** the user enters wrong confirmation words
- **THEN** the command aborts
- **AND** no slot is added
- **AND** the user is prompted to retry or cancel

### Requirement: Recovery setup output routing
`lango security recovery setup` SHALL write its mnemonic banner, written-down confirmation prompt, confirmation-word prompt, and success message through the Cobra command output stream so wrappers and test harnesses can capture non-error output without intercepting process-global stdout.

#### Scenario: Recovery setup output writes to command output
- **WHEN** the user runs `lango security recovery setup`
- **THEN** the mnemonic banner, written-down confirmation prompt, confirmation-word prompt, and success confirmation write to the Cobra command output stream

### Requirement: Mnemonic-based recovery

The system SHALL provide `lango security recovery restore` command that accepts a mnemonic, unwraps the MK from the matching mnemonic slot, and prompts the user for a new passphrase. The new passphrase replaces the existing passphrase slot via envelope re-wrap.

#### Scenario: Successful recovery with correct mnemonic

- **WHEN** the user runs `lango security recovery restore` and enters the correct mnemonic
- **THEN** the MK is unwrapped via `UnwrapFromMnemonic(mnemonic)`
- **AND** the user is prompted for a new passphrase
- **AND** `ChangePassphraseSlot(mk, newPassphrase)` updates the passphrase slot
- **AND** the envelope is persisted
- **AND** the user can open the application with the new passphrase on next launch

#### Scenario: Recovery with invalid mnemonic fails

- **WHEN** the user enters an invalid or wrong mnemonic
- **THEN** `UnwrapFromMnemonic` returns an error wrapping `ErrUnwrapFailed`
- **AND** the command reports the failure without modifying the envelope

### Requirement: Recovery restore output routing
`lango security recovery restore` SHALL write its success confirmation through the Cobra command output stream so wrappers and test harnesses can capture completion output without intercepting process-global stdout.

#### Scenario: Recovery restore success writes to command output
- **WHEN** the user runs `lango security recovery restore` and enters the correct mnemonic
- **THEN** the success confirmation writes to the Cobra command output stream

#### Scenario: Recovery restore warning output writes to command error stream
- **WHEN** `lango security recovery restore` emits keyfile or keyring update notices or warnings
- **THEN** those messages SHALL write to the Cobra command error stream

### Requirement: Mnemonic is never persisted

The mnemonic string SHALL only exist in memory during generation, display, and unwrap operations. It SHALL NOT be written to disk, logged, or stored in the envelope. Only the `WrappedMK`, `Salt`, and `Nonce` derived from the mnemonic are persisted.

#### Scenario: Mnemonic is zeroed after use

- **WHEN** `UnwrapFromMnemonic(mnemonic)` completes (success or failure)
- **THEN** any internal buffer containing the mnemonic is zeroed before the function returns
- **AND** no log message contains any portion of the mnemonic

### Requirement: Recovery is an explicit CLI action

Mnemonic recovery SHALL NOT be offered as an automatic prompt during bootstrap. Recovery is an explicit user action performed via `lango security recovery restore`. The restore command SHALL load the envelope directly from the filesystem without running the full bootstrap pipeline.

#### Scenario: Bootstrap does not prompt for mnemonic

- **WHEN** bootstrap Phase 4 (AcquireCredential) runs with an envelope containing a mnemonic slot
- **THEN** the mnemonic choice prompt SHALL NOT be shown
- **AND** passphrase acquisition SHALL proceed via the normal priority chain (KMS, keyring, keyfile, interactive, stdin)

#### Scenario: Mnemonic recovery via dedicated command without bootstrap

- **WHEN** the user runs `lango security recovery restore`
- **THEN** the command SHALL load the envelope directly via `security.LoadEnvelopeFile(langoDir)` without invoking the full bootstrap pipeline
- **AND** the user SHALL be prompted for the 24-word mnemonic
- **AND** on success, the user SHALL set a new passphrase via `ChangePassphraseSlot`

#### Scenario: Restore reports clear error when no envelope exists

- **WHEN** the user runs `lango security recovery restore` and no envelope file exists
- **THEN** the command SHALL return an error: `"envelope not found — recovery requires local encryption mode"`

#### Scenario: Non-interactive restore fails gracefully

- **WHEN** `lango security recovery restore` is run in a non-interactive environment
- **THEN** it SHALL return an error requiring an interactive terminal

### Requirement: Recovery confirmation-word prompt uses shared prompt helper
`lango security recovery setup` SHALL route its confirmation-word prompt through the shared visible line-entry prompt helper using Cobra command input/output streams.

#### Scenario: Recovery confirmation-word accepts matching final line without trailing newline
- **WHEN** the operator enters the correct confirmation word and the input stream ends immediately after that line without a trailing newline
- **THEN** `lango security recovery setup` SHALL accept the confirmation word instead of surfacing a read error

### Requirement: Recovery written-down confirmation treats EOF as denial
`lango security recovery setup` SHALL treat EOF on the written-down confirmation prompt as a clean setup abort.

#### Scenario: Recovery setup EOF aborts before word checks
- **WHEN** the written-down confirmation prompt reaches EOF before approval
- **THEN** the command SHALL abort setup
- **AND** it SHALL not proceed to the confirmation-word prompts

