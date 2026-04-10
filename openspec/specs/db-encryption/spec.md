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

The bootstrap sequence SHALL load the envelope file BEFORE acquiring a credential and opening the database when encryption is detected or enabled. When an envelope exists, the database encryption key SHALL be derived from the Master Key via `HKDF(mk, "lango-db-encryption")` and passed as `PRAGMA key = "x'<hex>'"` (raw key mode, no SQLCipher internal PBKDF2). Legacy installations (no envelope) SHALL continue to use the passphrase directly as `PRAGMA key = '<passphrase>'` until migration completes.

#### Scenario: Opening encrypted DB with envelope

- **WHEN** the DB is encrypted and an envelope exists with no pending flags
- **THEN** bootstrap unwraps the MK, derives the DB key via `DeriveDBKey(mk)`, hex-encodes it, and executes `PRAGMA key = "x'<hex>'"` after `sql.Open`
- **AND** `PRAGMA cipher_page_size` is set from config

#### Scenario: Opening encrypted DB with legacy format

- **WHEN** the DB is encrypted and no envelope exists
- **THEN** bootstrap uses the passphrase directly as `PRAGMA key = '<passphrase>'`
- **AND** the MigrateEnvelope phase will later convert to MK-derived key

#### Scenario: Opening plaintext DB

- **WHEN** the DB is not encrypted and `dbEncryption.enabled` is false
- **THEN** the database opens without any encryption PRAGMAs regardless of envelope state

#### Scenario: Pending migration or rekey uses legacy key fallback

- **WHEN** the DB is encrypted, envelope exists, and `PendingMigration` or `PendingRekey` is true
- **THEN** bootstrap opens the DB with the legacy passphrase key (not MK-derived)
- **AND** the MigrateEnvelope phase retries the pending operation after DB open

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

### Requirement: DB key derivation from Master Key

The system SHALL derive the SQLCipher database encryption key from the Master Key using HKDF-SHA256 with a domain-separated info label. The derivation SHALL produce a 32-byte raw key suitable for `PRAGMA key = "x'<hex>'"`. Key derivation SHALL be deterministic: the same MK always yields the same DB key.

#### Scenario: DeriveDBKey returns 32 bytes

- **WHEN** `DeriveDBKey(mk)` is called with a 32-byte MK
- **THEN** the returned byte slice has length 32

#### Scenario: DeriveDBKey is deterministic

- **WHEN** `DeriveDBKey(mk)` is called twice with the same MK
- **THEN** both calls return identical bytes

#### Scenario: Different MKs produce different DB keys

- **WHEN** `DeriveDBKey(mk1)` and `DeriveDBKey(mk2)` are called with different 32-byte MKs
- **THEN** the returned keys differ with overwhelming probability (cryptographic security)

#### Scenario: DB key uses domain separation

- **WHEN** HKDF is called to derive the DB key
- **THEN** the info parameter is exactly `"lango-db-encryption"` (byte-for-byte)
- **AND** this ensures the DB key is cryptographically independent from other keys derivable from the same MK

### Requirement: openDatabase accepts rawKey parameter

The `openDatabase` function SHALL accept a `rawKey bool` parameter to distinguish raw-key mode from passphrase mode. When `rawKey = true`, the function SHALL issue `PRAGMA key = "x'<encryptionKey>'"` where `encryptionKey` is a hex-encoded 32-byte key. When `rawKey = false`, the function SHALL issue `PRAGMA key = '<encryptionKey>'` where `encryptionKey` is a passphrase string.

#### Scenario: Raw key mode

- **WHEN** `openDatabase(path, hexKey, true, pageSize)` is called
- **THEN** the function executes `PRAGMA key = "x'<hexKey>'"` after `sql.Open`

#### Scenario: Passphrase mode

- **WHEN** `openDatabase(path, passphrase, false, pageSize)` is called
- **THEN** the function executes `PRAGMA key = '<passphrase>'` after `sql.Open`

#### Scenario: No encryption

- **WHEN** `openDatabase(path, "", false, 0)` is called with an empty encryption key
- **THEN** no `PRAGMA key` is executed
- **AND** the database is opened in plaintext mode

### Requirement: One-time PRAGMA rekey migration

The system SHALL perform a one-time `PRAGMA rekey` when migrating a legacy encrypted DB (passphrase key) to the envelope format (MK-derived raw key). The rekey SHALL execute after successful data re-encryption. The system SHALL back up the DB via `VACUUM INTO` before the rekey operation, with `PRAGMA wal_checkpoint(TRUNCATE)` preceding the backup to ensure WAL safety.

#### Scenario: Successful rekey on SQLCipher migration

- **WHEN** legacy migration runs on a SQLCipher DB
- **THEN** the system executes `PRAGMA wal_checkpoint(TRUNCATE)`, then `VACUUM INTO 'lango.db.pre-migration'`, then `PRAGMA rekey = "x'<HKDF(mk)>'"`, then closes and reopens the DB with the new raw key to verify
- **AND** on success, clears `PendingRekey` in the envelope and persists

#### Scenario: Rekey failure retains backup

- **WHEN** `PRAGMA rekey` fails or reopen verification fails
- **THEN** the `lango.db.pre-migration` backup file is retained
- **AND** the envelope keeps `PendingRekey = true`
- **AND** the next bootstrap retries or the user can restore from backup

#### Scenario: Plaintext DB migration skips rekey

- **WHEN** legacy migration runs on a plaintext DB
- **THEN** no `PRAGMA rekey` is executed
- **AND** `PendingRekey` is never set to true for this migration
