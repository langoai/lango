## Purpose

Capability spec for db-encryption. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: DB encryption configuration
The system MUST support a `security.dbEncryption` configuration with `enabled` (bool) and `cipherPageSize` (int, default 4096) fields.

#### Scenario: Default configuration
- **WHEN** no dbEncryption config is specified
- **THEN** `enabled` defaults to `false` and `cipherPageSize` defaults to `4096`

### Requirement: Encrypted DB detection
The system MUST detect whether a database file is encrypted by inspecting the first 16 bytes of the file header. Standard SQLite files start with "SQLite format 3\0"; encrypted files do not.

#### Scenario: Plaintext DB detection
- **WHEN** the DB file starts with "SQLite format 3"
- **THEN** `IsDBEncrypted()` returns `false`

#### Scenario: Encrypted DB detection
- **WHEN** the DB file does not start with "SQLite format 3"
- **THEN** `IsDBEncrypted()` returns `true`

#### Scenario: Non-existent DB
- **WHEN** the DB file does not exist
- **THEN** `IsDBEncrypted()` returns `false`

### Requirement: Bootstrap with encrypted DB
The bootstrap sequence MUST acquire the passphrase BEFORE opening the database when encryption is detected or enabled. The passphrase is passed as `PRAGMA key` followed by `PRAGMA cipher_page_size`.

#### Scenario: Opening encrypted DB
- **WHEN** the DB is encrypted or `dbEncryption.enabled` is true
- **THEN** the passphrase is acquired first, and `PRAGMA key` + `PRAGMA cipher_page_size` are executed after `sql.Open`

#### Scenario: Opening plaintext DB
- **WHEN** the DB is not encrypted and `dbEncryption.enabled` is false
- **THEN** the database opens without any encryption PRAGMAs

### Requirement: Plaintext to encrypted migration
`MigrateToEncrypted(dbPath, passphrase, cipherPageSize)` MUST convert a plaintext SQLite DB to SQLCipher format using `ATTACH DATABASE ... KEY` + `sqlcipher_export()`.

#### Scenario: Successful migration
- **WHEN** the source DB is plaintext and passphrase is non-empty
- **THEN** an encrypted copy is created, verified, atomically swapped, and the plaintext backup is securely deleted

#### Scenario: Already encrypted
- **WHEN** the source DB is already encrypted
- **THEN** the function returns an error without modifying the file

#### Scenario: Empty passphrase
- **WHEN** passphrase is empty
- **THEN** the function returns an error

### Requirement: Encrypted to plaintext decryption
`DecryptToPlaintext(dbPath, passphrase, cipherPageSize)` MUST convert a SQLCipher-encrypted DB back to plaintext using reverse `sqlcipher_export()`.

#### Scenario: Successful decryption
- **WHEN** the source DB is encrypted and correct passphrase is provided
- **THEN** a plaintext copy is created, verified, atomically swapped, and the encrypted backup is securely deleted

#### Scenario: Not encrypted
- **WHEN** the source DB is not encrypted
- **THEN** the function returns an error

### Requirement: CLI db-migrate command
`lango security db-migrate` MUST encrypt the application database. It requires interactive confirmation unless `--force` is used.

#### Scenario: Interactive migration
- **WHEN** the user runs `lango security db-migrate` in an interactive terminal
- **THEN** a confirmation prompt is shown before proceeding

#### Scenario: Non-interactive with --force
- **WHEN** the user runs `lango security db-migrate --force`
- **THEN** migration proceeds without confirmation

### Requirement: CLI db-decrypt command
`lango security db-decrypt` MUST decrypt the application database back to plaintext. Same confirmation behavior as db-migrate.

### Requirement: Security status display
`lango security status` MUST display the DB encryption state as one of: "encrypted (active)", "enabled (pending migration)", or "disabled (plaintext)".

#### Scenario: Encrypted DB
- **WHEN** the DB file is encrypted
- **THEN** status shows "encrypted (active)"

#### Scenario: Config enabled, DB plaintext
- **WHEN** `dbEncryption.enabled` is true but DB is not encrypted
- **THEN** status shows "enabled (pending migration)"

#### Scenario: Config disabled
- **WHEN** `dbEncryption.enabled` is false and DB is not encrypted
- **THEN** status shows "disabled (plaintext)"

### Requirement: Secure file deletion
Plaintext backup files MUST be overwritten with zeros before removal to prevent recovery from disk.
