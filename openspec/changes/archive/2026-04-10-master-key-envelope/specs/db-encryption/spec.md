## MODIFIED Requirements

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

## ADDED Requirements

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
