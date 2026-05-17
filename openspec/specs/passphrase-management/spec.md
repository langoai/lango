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

### Requirement: Change-passphrase success output routing
`lango security change-passphrase` SHALL write its non-error success confirmation through the Cobra command output stream so wrappers and test harnesses can capture completion output without intercepting process-global stdout.

#### Scenario: Change-passphrase success writes to command output
- **WHEN** `lango security change-passphrase` succeeds
- **THEN** the command writes the success confirmation to the Cobra command output stream

#### Scenario: Change-passphrase warning output writes to command error stream
- **WHEN** `lango security change-passphrase` emits keyfile or keyring update notices or warnings
- **THEN** those messages SHALL write to the Cobra command error stream

#### Scenario: Change-passphrase with wrong current passphrase

- **WHEN** the user enters an incorrect current passphrase
- **THEN** `UnwrapFromPassphrase` returns `ErrUnwrapFailed`
- **AND** the command displays an error and aborts without modifying the envelope

#### Scenario: Deprecated migrate-passphrase command

- **WHEN** the user runs `lango security migrate-passphrase`
- **THEN** the command displays a deprecation notice pointing to `change-passphrase`
- **AND** either delegates to change-passphrase or completes its legacy behavior for backward compatibility

### Requirement: Migrate-passphrase output routing
`lango security migrate-passphrase` SHALL write its non-error status and success output through the Cobra command output stream so wrappers and test harnesses can capture migration progress without intercepting process-global stdout.

#### Scenario: Migrate-passphrase progress writes to command output
- **WHEN** `lango security migrate-passphrase` runs
- **THEN** the command writes its migration guidance, progress, and success output to the Cobra command output stream

#### Scenario: Change-passphrase failure leaves envelope intact

- **WHEN** envelope re-wrap fails during file write
- **THEN** the original envelope file remains unchanged (atomic replace or temp-file-rename pattern)
- **AND** the user can retry with the original passphrase

### Requirement: Passphrase change updates stored credentials
Successful passphrase change or recovery restore SHALL update any stored keyring or keyfile credentials so local unlock paths remain consistent with the new passphrase.

#### Scenario: Keyring updated after passphrase change
- **WHEN** `lango security change-passphrase` succeeds
- **THEN** the command SHALL attempt to update the secure keyring with the new passphrase
- **AND** failure SHALL print a warning with manual fix instructions
- **AND** the manual fix instructions SHALL point to `lango security keyring store`
- **AND** the manual fix instructions SHALL NOT point to nonexistent keyring subcommands

#### Scenario: Keyfile updated after passphrase change
- **WHEN** `lango security change-passphrase` succeeds and a keyfile exists
- **THEN** the command SHALL write the new passphrase to the keyfile

#### Scenario: Recovery restore updates stored credentials
- **WHEN** `lango security recovery restore` succeeds
- **THEN** the same keyring and keyfile update logic SHALL apply as in passphrase change
- **AND** keyring update failure guidance SHALL point to `lango security keyring store`

### Requirement: Passphrase no longer directly encrypts data

With the envelope architecture, the passphrase SHALL function as a Key Encryption Key (KEK) source only. It SHALL NOT be used directly as a data encryption key. The Master Key (MK) is the sole data encryption key, and the passphrase-derived KEK is used only to wrap/unwrap the MK.

#### Scenario: Passphrase role after migration

- **WHEN** an envelope exists and bootstrap is running
- **THEN** the passphrase SHALL be used only to derive a KEK and unwrap the MK
- **AND** all `Encrypt`/`Decrypt` operations on the `CryptoProvider` SHALL use the MK (stored as `keys["local"]`)
- **AND** the raw passphrase SHALL NOT be accessible after bootstrap completes

#### Scenario: Legacy environment variable path remains unsupported
- **WHEN** a caller attempts to configure the passphrase via `LANGO_PASSPHRASE`
- **THEN** the system SHALL ignore that environment variable for local-crypto bootstrap
- **AND** continue using the documented acquisition chain instead

### Requirement: Keyring-clear confirmation uses shared command streams
`lango security keyring clear` SHALL drive its interactive confirmation through the shared confirmation helper using Cobra command input/output streams. When stdin is non-interactive and `--force` is not provided, the command SHALL refuse to continue with explicit `--force` guidance instead of attempting a prompt.

#### Scenario: Keyring-clear denial prints abort message
- **WHEN** `lango security keyring clear` is run interactively and the user answers `n`
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave stored credentials untouched

#### Scenario: Keyring-clear confirm prints prompt on command output
- **WHEN** `lango security keyring clear` is run interactively and the user answers `y`
- **THEN** the confirmation prompt SHALL be written through the Cobra command output stream
- **AND** the command SHALL continue with the backend clearing flow

#### Scenario: Keyring-clear non-interactive path requires force
- **WHEN** `lango security keyring clear` is run with non-interactive stdin and without `--force`
- **THEN** the command SHALL return an error directing the user to pass `--force for non-interactive deletion`

### Requirement: Keyring-clear EOF aborts cleanly
`lango security keyring clear` SHALL treat EOF on its confirmation input as a clean denial.

#### Scenario: Keyring-clear EOF aborts without clearing
- **WHEN** `lango security keyring clear` prompts for confirmation and stdin reaches EOF before approval
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL NOT clear stored credentials

### Requirement: Security passphrase prompts use command output streams
Security commands that prompt for current, new, or stored passphrases SHALL write visible passphrase prompt text through the Cobra command output stream instead of process-global stdout.

#### Scenario: Change-passphrase prompts use command output
- **WHEN** `lango security change-passphrase` prompts for the current passphrase, new passphrase, and confirmation
- **THEN** each visible hidden-input prompt SHALL be written through the Cobra command output stream

#### Scenario: Migrate-passphrase prompts use command output
- **WHEN** `lango security migrate-passphrase` prompts for a new passphrase and confirmation
- **THEN** each visible hidden-input prompt SHALL be written through the Cobra command output stream

#### Scenario: Keyring store prompt uses command output
- **WHEN** `lango security keyring store` prompts for the passphrase to store
- **THEN** the visible hidden-input prompt SHALL be written through the Cobra command output stream

#### Scenario: Warning output remains on command error stream
- **WHEN** passphrase-changing commands emit keyfile or keyring update notices or warnings
- **THEN** those messages SHALL continue to write through the Cobra command error stream

### Requirement: Security passphrase command guards use command input streams
Interactive `lango security` passphrase commands SHALL validate the Cobra command input stream before hidden passphrase prompts instead of reading process-global stdin directly.

#### Scenario: Change-passphrase guard uses command input
- **WHEN** `lango security change-passphrase` reaches its interactive guard
- **THEN** the guard SHALL validate the command input stream

#### Scenario: Migrate-passphrase guard uses command input
- **WHEN** `lango security migrate-passphrase` reaches its interactive guard
- **THEN** the guard SHALL validate the command input stream

#### Scenario: Keyring-store guard uses command input
- **WHEN** `lango security keyring store` reaches its interactive guard
- **THEN** the guard SHALL validate the command input stream
