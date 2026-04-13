## ADDED Requirements

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

### Requirement: Mnemonic is never persisted

The mnemonic string SHALL only exist in memory during generation, display, and unwrap operations. It SHALL NOT be written to disk, logged, or stored in the envelope. Only the `WrappedMK`, `Salt`, and `Nonce` derived from the mnemonic are persisted.

#### Scenario: Mnemonic is zeroed after use

- **WHEN** `UnwrapFromMnemonic(mnemonic)` completes (success or failure)
- **THEN** any internal buffer containing the mnemonic is zeroed before the function returns
- **AND** no log message contains any portion of the mnemonic

### Requirement: Recovery during bootstrap

During bootstrap, if an envelope exists and contains at least one mnemonic slot, the system SHALL offer the user a choice between passphrase and mnemonic credential acquisition. If the user selects mnemonic, the passphrase acquisition is skipped and the MK is unwrapped directly from the mnemonic slot.

#### Scenario: Bootstrap offers recovery option when mnemonic slot exists

- **WHEN** bootstrap Phase 4 (AcquireCredential) runs with an envelope containing a mnemonic slot
- **AND** the terminal is interactive
- **THEN** the user is prompted to choose between passphrase and mnemonic
- **AND** selecting mnemonic triggers `UnwrapFromMnemonic` and sets `RecoveryMode = true`

#### Scenario: Non-interactive bootstrap skips recovery option

- **WHEN** bootstrap runs in a non-interactive environment (no tty, keyfile available)
- **THEN** the recovery choice prompt is skipped
- **AND** the keyfile passphrase path is used
