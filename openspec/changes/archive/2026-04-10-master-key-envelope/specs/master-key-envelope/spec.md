## ADDED Requirements

### Requirement: Master Key generation

The system SHALL generate a 32-byte Master Key (MK) from a cryptographically secure random source using `crypto/rand`. The MK is the root of data encryption and DB encryption key derivation.

#### Scenario: Fresh install generates new MK

- **WHEN** bootstrap runs with no existing envelope file
- **THEN** the system calls `GenerateMasterKey()` which returns 32 random bytes
- **AND** the MK is wrapped with a passphrase-derived KEK and stored in `envelope.json`
- **AND** the raw MK is held in memory only for the duration of the session

#### Scenario: MK is never written to disk in plaintext

- **WHEN** the MK is generated or loaded
- **THEN** the MK SHALL NOT be logged, serialized, or written to any file in plaintext
- **AND** the MK SHALL always be wrapped by at least one KEK slot before persistence

### Requirement: Master Key envelope structure

The system SHALL store the MK in a `MasterKeyEnvelope` structure with version, KEK slots, crash recovery flags, and timestamps. The envelope SHALL be serialized as JSON and stored at `<LangoDir>/envelope.json` with `0600` permissions.

#### Scenario: Envelope contains required fields

- **WHEN** a new envelope is created via `NewEnvelope(passphrase)`
- **THEN** the envelope has `Version = 1`
- **AND** at least one `KEKSlot` of type `passphrase`
- **AND** `PendingMigration = false` and `PendingRekey = false` (except during migration)
- **AND** `CreatedAt` and `UpdatedAt` timestamps set to current time

#### Scenario: Envelope file has restrictive permissions

- **WHEN** `StoreEnvelopeFile(langoDir, envelope)` is called
- **THEN** the file is created with `0600` permissions (owner read/write only)
- **AND** parent directory is ensured to exist

### Requirement: KEK slot structure and metadata

Each KEK slot SHALL include KDF algorithm metadata (`KDFAlg`, `KDFParams`), wrap algorithm (`WrapAlg`), domain separation string (`Domain`), slot-specific salt, wrapped MK ciphertext, GCM nonce, creation timestamp, optional label, and a stable UUID identifier. This metadata enables future algorithm migration without breaking existing slots.

#### Scenario: New passphrase slot has KDF metadata

- **WHEN** a passphrase KEK slot is added to the envelope
- **THEN** `KDFAlg = "pbkdf2-sha256"`
- **AND** `KDFParams.Iterations = 100000`
- **AND** `WrapAlg = "aes-256-gcm"`
- **AND** `Domain = "passphrase"`
- **AND** `ID` is a valid UUID string
- **AND** `Salt` is 16 random bytes unique to this slot

#### Scenario: Slots with different KDFs coexist

- **WHEN** the envelope has a passphrase slot (`pbkdf2-sha256`) and a future Argon2id slot (`argon2id`)
- **THEN** `DeriveKEK(secret, slot)` dispatches on `slot.KDFAlg`
- **AND** each slot unwraps the same MK independently

### Requirement: MK wrapping and unwrapping

The system SHALL wrap the MK using AES-256-GCM with the KEK as the key and a random 12-byte nonce. The ciphertext and nonce are stored in the slot's `WrappedMK` and `Nonce` fields. Unwrapping SHALL verify the GCM authentication tag before returning plaintext.

#### Scenario: Wrap-unwrap round trip

- **WHEN** `WrapMasterKey(mk, kek)` produces `(wrapped, nonce)` and `UnwrapMasterKey(wrapped, nonce, kek)` is called with the same KEK
- **THEN** the returned bytes equal the original MK

#### Scenario: Unwrap with wrong KEK fails

- **WHEN** `UnwrapMasterKey(wrapped, nonce, wrongKEK)` is called
- **THEN** the function returns an error wrapping `ErrUnwrapFailed`
- **AND** no plaintext is returned

#### Scenario: Unwrap with tampered ciphertext fails

- **WHEN** any byte of `wrapped` is modified and `UnwrapMasterKey` is called
- **THEN** the GCM authentication check fails and an error is returned

### Requirement: Multi-slot wrap/unwrap operations

The envelope SHALL support multiple KEK slots wrapping the same MK independently. The system SHALL provide `AddSlot`, `RemoveSlot`, `UnwrapFromPassphrase`, `UnwrapFromMnemonic`, and `ChangePassphraseSlot` operations. The system SHALL reject removal of the last slot.

#### Scenario: Add mnemonic slot to passphrase-only envelope

- **WHEN** an envelope has only a passphrase slot and `AddSlot(KEKSlotMnemonic, ...)` is called with the unwrapped MK
- **THEN** the envelope has 2 slots
- **AND** both slots unwrap the same MK
- **AND** passphrase and mnemonic independently produce the MK

#### Scenario: Remove last slot is rejected

- **WHEN** the envelope has only one slot and `RemoveSlot(slotID)` is called
- **THEN** the function returns `ErrLastSlot`
- **AND** the envelope is unchanged

#### Scenario: Change passphrase without touching other slots

- **WHEN** `ChangePassphraseSlot(mk, newPassphrase)` is called
- **THEN** the passphrase slot's `Salt`, `WrappedMK`, and `Nonce` are replaced
- **AND** mnemonic slots (if any) are unchanged
- **AND** data in the DB remains untouched

### Requirement: DB key derivation from MK

The system SHALL derive the SQLCipher database key from the MK using HKDF-SHA256 with a domain-separated info label `"lango-db-encryption"`. The derived key is 32 bytes, hex-encoded, and used as `PRAGMA key = "x'<hex>'"` (raw key mode, no SQLCipher internal PBKDF2).

#### Scenario: DB key is deterministic from MK

- **WHEN** `DeriveDBKey(mk)` is called twice with the same MK
- **THEN** both calls return the same 32-byte key

#### Scenario: DB key changes with MK

- **WHEN** `DeriveDBKey(mk1)` and `DeriveDBKey(mk2)` are called with different MKs
- **THEN** the returned keys are different with overwhelming probability

#### Scenario: Passphrase change does not change DB key

- **WHEN** the passphrase is changed via envelope re-wrap
- **THEN** the MK is unchanged
- **AND** `DeriveDBKey(mk)` returns the same value
- **AND** no `PRAGMA rekey` is needed

### Requirement: Crash recovery flags

The envelope SHALL track migration state with `PendingMigration` and `PendingRekey` boolean flags. These flags SHALL be set to `true` before data re-encryption or `PRAGMA rekey`, and cleared to `false` only after successful completion.

#### Scenario: PendingMigration set before data re-encryption

- **WHEN** migration begins
- **THEN** the envelope is stored with `PendingMigration = true`
- **AND** the envelope file write completes before SQL transactions start

#### Scenario: PendingRekey cleared only after successful rekey

- **WHEN** `PRAGMA rekey` succeeds and the DB is reopened with the new key
- **THEN** the envelope is updated with `PendingRekey = false`
- **AND** the updated envelope is persisted to disk

#### Scenario: Crash during migration leaves consistent state

- **WHEN** the process crashes after `StoreEnvelopeFile` (PendingMigration=true) but before the SQL transaction commits
- **THEN** the next bootstrap detects `PendingMigration = true`
- **AND** retries data re-encryption using the legacy passphrase-derived key

### Requirement: Exported ZeroBytes utility

The system SHALL provide an exported `ZeroBytes(b []byte)` function that overwrites all bytes with zero. This function replaces scattered private `zeroBytes` copies in `wallet` and `p2p` packages.

#### Scenario: ZeroBytes clears the slice

- **WHEN** `ZeroBytes(buf)` is called on a non-empty byte slice
- **THEN** all bytes of `buf` are `0x00`
