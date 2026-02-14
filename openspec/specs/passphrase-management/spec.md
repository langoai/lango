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
The system SHALL store a checksum to detect incorrect passphrase early.

#### Scenario: Checksum storage
- **WHEN** new passphrase is set
- **THEN** system stores SHA256(passphrase + salt) in security_config table

#### Scenario: Checksum verification
- **WHEN** passphrase is entered
- **THEN** system computes SHA256(passphrase + salt)
- **AND** compares with stored checksum
- **AND** rejects if mismatch before attempting any decryption

### Requirement: Passphrase Migration Command
The system SHALL provide a CLI command to migrate encrypted data to a new passphrase.

#### Scenario: Successful migration
- **WHEN** user runs `lango security migrate-passphrase`
- **AND** enters correct current passphrase
- **AND** enters new passphrase (with confirmation)
- **THEN** system decrypts all secrets with old key
- **AND** re-encrypts with new key
- **AND** updates salt and checksum

#### Scenario: Migration with wrong current passphrase
- **WHEN** user enters incorrect current passphrase
- **THEN** system displays error and aborts without modifying data

#### Scenario: Migration rollback on failure
- **WHEN** re-encryption fails mid-process
- **THEN** system rolls back all changes
- **AND** preserves original encrypted data

