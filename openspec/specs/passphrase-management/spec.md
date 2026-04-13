# Passphrase Management

## Purpose
This capability defines how the user's passphrase for the Local Crypto Provider is securely handled, validated, and migrated. It ensures passphrases are never stored in plain text configuration files and provides mechanisms for key rotation.

## Requirements

### Requirement: Passphrase source resolution
The system SHALL resolve passphrases using the priority chain: keyfile (`~/.lango/keyfile`) → interactive terminal prompt → stdin pipe. The system SHALL NOT read passphrases from the `LANGO_PASSPHRASE` environment variable or the `security.passphrase` config field.

#### Scenario: Passphrase acquisition in CLI security commands
- **WHEN** `initLocalCrypto` is called in CLI security commands
- **THEN** the passphrase is acquired via `passphrase.Acquire()` (not env var or config)

#### Scenario: Non-interactive environment without keyfile
- **WHEN** stdin is not a terminal and no keyfile exists
- **THEN** the system attempts to read from stdin pipe; if empty, returns an error

### Requirement: Passphrase Checksum Validation

The system SHALL validate passphrase correctness. For envelope-based installations, validation SHALL occur via `UnwrapFromPassphrase` which verifies the AES-GCM authentication tag on the wrapped MK. For legacy installations (no envelope), the system SHALL continue to use the HMAC-SHA256 checksum stored in `security_config` until migration completes.

#### Scenario: Envelope-based passphrase verification

- **WHEN** a bootstrap loads an envelope and calls `envelope.UnwrapFromPassphrase(passphrase)`
- **THEN** the passphrase is verified implicitly via AES-GCM authentication
- **AND** a wrong passphrase returns `ErrUnwrapFailed` without revealing which slot was attempted

#### Scenario: Legacy checksum verification during migration

- **WHEN** a bootstrap detects legacy format (salt and checksum exist, no envelope)
- **THEN** the system computes `HMAC-SHA256(passphrase, salt)` and compares with stored checksum
- **AND** rejects if mismatch before attempting any decryption
- **AND** proceeds to migration if checksum matches

#### Scenario: Legacy checksum stays in security_config after migration

- **WHEN** migration to envelope completes
- **THEN** the legacy `security_config.default` row (salt + checksum) SHALL remain in the DB as a downgrade safety artifact
- **AND** it SHALL NOT be consulted during subsequent envelope-based bootstrap

### Requirement: Passphrase Migration Command

The system SHALL provide a CLI command to change the passphrase. For envelope-based installations, the command SHALL re-wrap the existing MK with a new KEK derived from the new passphrase — no data re-encryption and no DB rekey. The legacy `lango security migrate-passphrase` command SHALL be deprecated in favor of `lango security change-passphrase`.

#### Scenario: Change passphrase on envelope-based install

- **WHEN** the user runs `lango security change-passphrase`
- **AND** enters the correct current passphrase
- **AND** enters a new passphrase (with confirmation)
- **THEN** the system unwraps the MK from the passphrase slot
- **AND** calls `envelope.ChangePassphraseSlot(mk, newPassphrase)` which generates a new salt, derives a new KEK, and re-wraps the MK
- **AND** persists the updated envelope via `StoreEnvelopeFile`
- **AND** does NOT re-encrypt any secrets or config_profiles rows
- **AND** does NOT call `PRAGMA rekey`

#### Scenario: Change-passphrase with wrong current passphrase

- **WHEN** the user enters an incorrect current passphrase
- **THEN** `UnwrapFromPassphrase` returns `ErrUnwrapFailed`
- **AND** the command displays an error and aborts without modifying the envelope

#### Scenario: Deprecated migrate-passphrase command

- **WHEN** the user runs `lango security migrate-passphrase`
- **THEN** the command displays a deprecation notice pointing to `change-passphrase`
- **AND** either delegates to change-passphrase or completes its legacy behavior for backward compatibility

#### Scenario: Change-passphrase failure leaves envelope intact

- **WHEN** envelope re-wrap fails during file write
- **THEN** the original envelope file remains unchanged (atomic replace or temp-file-rename pattern)
- **AND** the user can retry with the original passphrase

### Requirement: Passphrase change updates stored credentials

#### Scenario: Keyring updated after passphrase change
- **WHEN** `lango security change-passphrase` succeeds
- **THEN** the command SHALL attempt to update the secure keyring with the new passphrase
- **AND** failure SHALL print a warning with manual fix instructions

#### Scenario: Keyfile updated after passphrase change
- **WHEN** `lango security change-passphrase` succeeds and a keyfile exists
- **THEN** the command SHALL write the new passphrase to the keyfile

#### Scenario: Recovery restore updates stored credentials
- **WHEN** `lango security recovery restore` succeeds
- **THEN** the same keyring and keyfile update logic SHALL apply as in passphrase change

### Requirement: Passphrase no longer directly encrypts data

With the envelope architecture, the passphrase SHALL function as a Key Encryption Key (KEK) source only. It SHALL NOT be used directly as a data encryption key. The Master Key (MK) is the sole data encryption key, and the passphrase-derived KEK is used only to wrap/unwrap the MK.

#### Scenario: Passphrase role after migration

- **WHEN** an envelope exists and bootstrap is running
- **THEN** the passphrase SHALL be used only to derive a KEK and unwrap the MK
- **AND** all `Encrypt`/`Decrypt` operations on the `CryptoProvider` SHALL use the MK (stored as `keys["local"]`)
- **AND** the raw passphrase SHALL NOT be accessible after bootstrap completes

