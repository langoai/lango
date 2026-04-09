## Purpose

Capability spec for keyfile-shred. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Secure keyfile shredding
The system SHALL provide a `ShredKeyfile()` function that securely destroys a passphrase keyfile by overwriting its content with zero bytes, syncing to disk, and removing the file. The operation SHALL be idempotent.

#### Scenario: Shred existing keyfile
- **WHEN** `ShredKeyfile()` is called with a path to an existing file
- **THEN** the file content is overwritten with zero bytes, synced to disk via `Sync()`, and the file is removed

#### Scenario: Shred nonexistent keyfile
- **WHEN** `ShredKeyfile()` is called with a path to a file that does not exist
- **THEN** the function returns nil without error

#### Scenario: Shred failure
- **WHEN** `ShredKeyfile()` encounters an I/O error during overwrite, sync, or removal
- **THEN** a wrapped error is returned describing the failed operation
