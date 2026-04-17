# Master Key Envelope

## Purpose

Defines the MK/KEK three-layer key hierarchy for Lango's local storage encryption. The Master Key (MK) is a random 32-byte root key wrapped by one or more Key Encryption Keys (KEK). KEKs are derived from user secrets (passphrase, recovery mnemonic) via PBKDF2. The envelope is stored as a JSON file alongside the database, enabling passphrase change without data re-encryption, recovery mnemonic support, and broker-managed payload protection key derivation.

## Requirements

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

#### Scenario: KMS KEK slot wraps MK via CryptoProvider
- **WHEN** `AddKMSSlot(ctx, label, mk, provider, kmsProviderName, kmsKeyID)` is called
- **THEN** the envelope SHALL call `provider.Encrypt(ctx, kmsKeyID, mk)` and store the returned ciphertext in `slot.WrappedMK`
- **AND** the slot SHALL have `Type: "hardware"`, `KDFAlg: "none"`, `WrapAlg: "kms-envelope"`
- **AND** `KMSProvider` and `KMSKeyID` SHALL be populated from the arguments

#### Scenario: KMS KEK slot fields use omitempty
- **WHEN** an envelope is serialized to JSON
- **THEN** `KMSProvider` and `KMSKeyID` fields SHALL use `omitempty` tags
- **AND** envelopes without KMS slots SHALL serialize identically to the pre-KMS format

### Requirement: MK unwrap from KMS with 2-tier matching

The `UnwrapFromKMS(ctx, provider, providerName, keyID)` method SHALL apply a two-tier matching strategy to find and decrypt the correct KMS slot.

#### Scenario: Tier 1 exact match (provider + keyID)
- **WHEN** `UnwrapFromKMS` is called and a slot exists with `KMSProvider == providerName && KMSKeyID == keyID`
- **THEN** the method SHALL call `provider.Decrypt(ctx, slot.KMSKeyID, slot.WrappedMK)`
- **AND** return the recovered MK on success

#### Scenario: Tier 2 provider-only fallback
- **WHEN** no Tier 1 exact match succeeds and slots exist with `KMSProvider == providerName`
- **THEN** the method SHALL try each matching slot using `slot.KMSKeyID` for the decrypt call (not the env keyID)
- **AND** return the recovered MK on first success

#### Scenario: No matching slot
- **WHEN** no `KEKSlotHardware` slot matches the configured provider
- **THEN** `UnwrapFromKMS` SHALL return `ErrKMSSlotUnavailable`

#### Scenario: Recovered MK size validation
- **WHEN** `provider.Decrypt` returns successfully
- **THEN** the recovered MK SHALL be validated to be exactly 32 bytes
- **AND** if the size is wrong, the method SHALL zero the bytes and return `ErrUnwrapFailed`

#### Scenario: KMS slot backward compatibility
- **WHEN** an envelope JSON without `kms_provider` or `kms_key_id` fields is loaded
- **THEN** those fields SHALL default to empty strings
- **AND** the envelope SHALL load and function correctly for passphrase/mnemonic slots

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

### Requirement: Envelope remains payload-protection root
The master-key envelope MUST remain the root source of key material for broker-managed payload protection after SQLCipher page encryption is removed.

#### Scenario: Envelope-backed payload protection
- **WHEN** the broker needs key material for payload encryption or decryption
- **THEN** it derives the required key material from the envelope-managed master key
- **AND** it does not derive or apply a SQLCipher page key

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
