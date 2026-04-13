## Context

Lango's current local storage encryption structure derives a key directly in `internal/security/local_provider.go` via `passphrase → PBKDF2(100k, SHA256) → AES-256-GCM key`. This key serves as the encryption key for both the `secrets` table and `config_profiles` table, and also as the SQLCipher key for `~/.lango/lango.db` through `session/ent_store.go`.

This structure is simple but has 3 structural problems:
1. **Passphrase loss = total loss**: no recovery path
2. **Passphrase change = full re-encryption**: all secrets and config_profiles must be re-encrypted with a new key
3. **Storage root and identity root are mixed**: after Phase 1, the storage side becomes a blocker when separating identity in Phase 2-4 (algorithm agility, DID v2, hybrid handshake)

Additionally, there is an inconsistency between `bootstrap.go:116` and `session/ent_store.go:99-102` in how the SQLCipher key is passed (`PRAGMA key = '<passphrase>'` vs `PRAGMA key = "x'<hex>'"`). This change unifies them to raw key mode.

**Stakeholders**: CLI users (passphrase change/recovery UX), CryptoProvider consumers (interface remains unchanged), future Phase 2-7 PQC work.

## Goals / Non-Goals

**Goals:**
- Random Master Key (MK) generation, wrap/unwrap support via multiple KEK slots
- No data re-encryption needed on passphrase change (envelope re-wrap only)
- BIP39 24-word recovery mnemonic slot support
- DB encryption key = HKDF(MK) for complete separation from passphrase
- Safe recovery on crash during migration (PendingMigration + PendingRekey flags)
- CryptoProvider interface remains unchanged — zero consumer impact
- `lango security status` passphrase-free default behavior (spec compliant)
- Delta updates reflected in 9 OpenSpec specs

**Non-Goals:**
- PQC algorithm introduction (Phase 2-3 scope)
- Identity root separation (Phase 0, 3 scope)
- Hardware token slot implementation (`KEKSlotHardware` type defined only, actual implementation is follow-up)
- `recovery_file` slot implementation (`KEKSlotRecoveryFile` type defined only, actual implementation is follow-up)
- Argon2id KDF support (KDFAlg metadata field prepared only, actual implementation is follow-up)
- Linking `LangoDir` as parent directory of `DBPath` (outside Phase 1 scope)
- Mnemonic recovery automation / headless environment support (interactive only)

## Decisions

### D1: Envelope storage location — filesystem (`~/.lango/envelope.json`)

**Choice**: `~/.lango/envelope.json` (JSON, 0600 permissions)

**Alternatives**:
- (A) DB `security_config` table row — When SQLCipher is active, losing the passphrase means the DB itself cannot be opened, making the envelope inaccessible. Mnemonic recovery becomes impossible.
- (B) OS keyring — Increases platform dependency, difficult to test.

**Rationale**: The only way to resolve the SQLCipher + recovery path contradiction in (A) is to place the envelope outside the DB. The JSON file is safe to store as plaintext since WrappedMK is encrypted with KEK (same design as LUKS, KeePass).

### D2: DB key derivation — `HKDF(MK, "lango-db-encryption")` raw 32-byte

**Choice**: `DeriveDBKey(mk) = HKDF-Expand(SHA256, mk, "lango-db-encryption", 32)`, hex encoded then `PRAGMA key = "x'<hex>'"`

**Alternatives**:
- (A) Direct passphrase use (current) — requires `PRAGMA rekey` on passphrase change.
- (B) Separate random DB key wrapped with MK — requires additional envelope field, more complex implementation.
- (C) `HKDF(MK, ...)` — If MK is immutable, DB key is immutable, completely independent from passphrase change.

**Rationale**: (C) is the simplest. Since MK is the sole root, all keys derived from MK are automatically stable. Passphrase change = envelope re-wrap only, zero DB impact.

### D3: Migration crash safety — `PendingMigration` + `PendingRekey` dual-flag

**Choice**: Two boolean flags `PendingMigration` and `PendingRekey` in the envelope. Save the envelope first (flag=true) before data re-encryption. Clear each flag upon stage completion.

**Alternatives**:
- (A) Single flag — cannot distinguish state on rekey failure.
- (B) SQL transaction only — `PRAGMA rekey` is not transactional.
- (C) Separate "migration_state" file — redundant state management.

**Rationale**: Two flags precisely represent the intermediate state of "data migration done, rekey pending". Bootstrap Phase 7 reads the flags directly and retries. Plaintext DB also uses PendingMigration for the same processing path.

### D4: DB backup — `PRAGMA wal_checkpoint(TRUNCATE)` + `VACUUM INTO`

**Choice**: `PRAGMA wal_checkpoint(TRUNCATE)` followed by `VACUUM INTO 'lango.db.pre-migration'`

**Alternatives**:
- (A) `cp lango.db lango.db.pre-rekey` — Risk of missing `-wal` file contents in WAL mode.
- (B) SQLite backup API — CGO call, complex.
- (C) `VACUUM INTO` — Creates a consistent snapshot regardless of WAL state.

**Rationale**: (C) is WAL-safe and operates at the SQL level. `wal_checkpoint` beforehand provides additional safety.

### D5: Bootstrap pipeline — 7 phases → 10 phases

**Choice**:
```
1. EnsureDataDir (modified)  2. DetectEncryption  3. LoadEnvelopeFile (new)
4. AcquireCredential (modified)  5. UnwrapOrCreateMK (new)  6. OpenDatabase (modified)
7. MigrateEnvelope (new)  8. LoadSecurityState (modified)  9. InitCrypto (modified)
10. LoadProfile
```

**Alternatives**:
- (A) Keep existing 7 phases with internal branching — phase responsibilities become blurred, harder to test.
- (B) Single large phase — simple but difficult to manage cleanup/restart.

**Rationale**: Each phase has a single responsibility. LoadEnvelopeFile executes before DB access, enabling recovery option assessment. MigrateEnvelope as a separate phase makes crash retry logic clear.

### D6: CryptoProvider interface remains unchanged

**Choice**: No changes to `Sign/Encrypt/Decrypt` signatures. Only add new methods to `LocalCryptoProvider`: `InitializeWithEnvelope`, `InitializeNewEnvelope`, `Envelope()`, `IsLegacy()`, `Close()`.

**Alternatives**:
- (A) Add KEK management methods to the interface — impacts all implementations (KMS, Composite, RPC).
- (B) Define new `EnvelopeCryptoProvider` interface — complex wiring, existing consumer changes.

**Rationale**: Consumer stability is the top priority. MK is stored in `keys["local"]`, so existing Encrypt/Decrypt/Sign paths are used as-is. KMS providers are unrelated to this change.

### D7: BIP39 library — `tyler-smith/go-bip39`

**Choice**: `github.com/tyler-smith/go-bip39` (confirmed by user)

**Alternatives**:
- (A) `cosmos/go-bip39` — Cosmos ecosystem oriented, excessive.
- (B) Custom implementation — wordlist management burden, bug risk.

**Rationale**: tyler-smith/go-bip39 is a Go ecosystem standard. Lightweight and single-purpose. Independent from go-ethereum.

### D8: Unified `security_config` access — `SecurityConfigStore`

**Choice**: `SecurityConfigStore` struct in `internal/security/config_store.go`. Delegate duplicated `ensureSecurityTable`/`loadSalt`/`storeSalt`/`loadChecksum`/`storeChecksum` from bootstrap.go and session/ent_store.go to this struct.

**Alternatives**:
- (A) Keep duplication — increased drift risk.
- (B) Model as Ent entity — conflicts with existing raw SQL pattern.

**Rationale**: Code currently duplicated in 2 places would grow to 3+ places with this change. Converging to a single store is the right approach for maintainability.

### D9: `lango security status` default behavior — passphrase-free non-interactive mini-bootstrap

**Choice**: Direct envelope file reading + `passphrase.AcquireNonInteractive()` (keyring/keyfile only, no prompt) + `openDatabaseReadOnly()` (read-only, no schema migration).

**Alternatives**:
- (A) Call existing `bootLoader()` — shows prompt in interactive environments, spec violation.
- (B) DB access only with `--full` flag — violates existing spec's "default passphrase-free" requirement.

**Rationale**: The current `cli-security-status` spec requires the default behavior to show all existing fields without a passphrase and degrade to zero counts when DB access fails. Non-interactive mini-bootstrap is required to comply. Reference: recent commit `8112bcc6`'s sandbox graceful degradation pattern.

### D10: Add `rawKey bool` to `openDatabase()` signature

**Choice**: `openDatabase(dbPath, encryptionKey string, rawKey bool, cipherPageSize int)`

**Alternatives**:
- (A) Separate openDatabase for raw key only — maintenance burden of two functions.
- (B) Prefix in `encryptionKey` (`"raw:..."`) — fragile, error-prone.

**Rationale**: `rawKey bool` parameter is explicit. Same code path handles only the `PRAGMA key = '...'` vs `PRAGMA key = "x'...'"` branching.

## Risks / Trade-offs

- **[Critical] `PRAGMA rekey` failure → split state** → `PendingRekey` flag + dual-open fallback (Phase 6: fallback to legacy passphrase if MK-derived key fails) + `VACUUM INTO` backup preservation
- **[Critical] MK exposed in log/error messages** → MK must never be included in error strings. Use only `fmt.Errorf("... %w", err)` wrapping. Verify with grep during code review
- **[High] Crash during migration data re-encrypt (plaintext)** → `PendingMigration` flag enables retry on next boot
- **[High] Crash between migration TX and rekey (SQLCipher)** → Saved as PendingMigration=false + PendingRekey=true. On next boot, open DB with passphrase (legacy key) → retry rekey
- **[High] `config_profiles` re-encryption missed** → COUNT verification inside migration TX (row count must match before and after)
- **[Medium] Envelope file permission issues** → Created with 0600. Permission check on load with warning only (no forced rejection)
- **[Medium] **Residual**: PendingMigration/PendingRekey=true + passphrase lost + only mnemonic exists** → MK can be unwrapped via mnemonic, but DB uses legacy passphrase key so it cannot be opened. Recovery from `VACUUM INTO` backup (`lango.db.pre-migration`) required. This scenario is the intersection of (migration crash) and (passphrase loss), making it extremely rare. Documented in docs + CLI warning
- **[Low] BIP39 dependency vulnerability** → `tyler-smith/go-bip39` version pinned, indirect dependency audit
- **[Low] Existing `migrate-passphrase` user confusion** → deprecation message + `change-passphrase` guidance
- **[Low] PBKDF2 parameter upgrade (future)** → Iteration stored in `KEKSlot.KDFParams`. Future new slots can use higher values; existing slots remain unchanged

## Migration Plan

### Phase 1a: Fresh install (first-run)
- `EnsureDataDir` → `LoadEnvelopeFile` (nil) → `AcquireCredential` (passphrase) → `UnwrapOrCreateMK` (create new envelope) → `OpenDatabase` (MK-derived key or plaintext) → `InitCrypto`
- User perspective: same passphrase prompt as before

### Phase 1b: Existing install upgrade (legacy → envelope)
- `EnsureDataDir` → `LoadEnvelopeFile` (nil) → `AcquireCredential` (passphrase) → `UnwrapOrCreateMK` (LegacyMode=true, skip) → `OpenDatabase` (passphrase as legacy key) → `MigrateEnvelope`:
  1. Derive old key + checksum verification
  2. Generate MK, create envelope (PendingMigration=true, PendingRekey=true if SQLCipher)
  3. `StoreEnvelopeFile()` — save envelope first
  4. `wal_checkpoint(TRUNCATE)` + `VACUUM INTO 'lango.db.pre-migration'` — WAL-safe backup
  5. SQL TX: re-encrypt secrets + config_profiles (with COUNT verification)
  6. PendingMigration=false, update envelope
  7. (SQLCipher) `PRAGMA rekey = "x'<HKDF(MK)>'"` + close/reopen verification
  8. PendingRekey=false, update envelope
  9. (Optional) Remove backup file
- User perspective: "Upgrading encryption format..." message on first boot. Normal operation afterwards.

### Phase 1c: Crash recovery
- Recovery per crash point is handled by Bootstrap Phase 7 (MigrateEnvelope) which reads `s.Envelope.PendingMigration`/`PendingRekey` directly and retries.
- Automatically completes on next boot.

### Rollback
- Phase 1 implementation itself is not rolled back (on downgrade, existing legacy paths cannot ignore the envelope).
- In emergency, restore `lango.db.pre-migration` backup file to `lango.db` + delete `envelope.json` → works with existing binary.
- Documentation needed: backup/restore procedure in `docs/security/envelope-migration.md`.

## Open Questions

- **Q1**: `SecurityConfigStore` consolidation scope — only clean up bootstrap.go and session/ent_store.go duplication, or include `MigrateSecrets` too? **Answer**: This change covers deduplication only. `MigrateSecrets` has internal logic and is a separate follow-up.
- **Q2**: Envelope file permission verification — reject if not 0600, or warning only? **Answer**: Warning only in Phase 1. Rejection is a follow-up when tighter security posture is needed.
- **Q3**: Recovery mnemonic passphrase protection (BIP39 optional passphrase) — support in Phase 1? **Answer**: Not supported. BIP39 optional passphrase increases UX complexity. Follow-up.
