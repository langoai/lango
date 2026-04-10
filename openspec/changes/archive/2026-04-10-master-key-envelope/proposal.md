## Why

Lango's current local storage encryption uses a `passphrase → PBKDF2 → AES-256-GCM key` structure where the passphrase directly serves as both the data key and the SQLCipher DB key. This causes 3 structural problems: (1) total data loss on passphrase loss, (2) full re-encryption required on passphrase change, and (3) no recovery path. This is the first step of Phase 1 Security & Crypto Renewal, separating the storage root from the identity root and establishing the key hierarchy foundation for future PQC transition.

## What Changes

- **New**: Master Key Envelope (MK/KEK) architecture introduction
  - Random 32-byte Master Key (MK) generation, stored in AES-256-GCM wrapped form in `~/.lango/envelope.json`
  - Passphrase → KEK (PBKDF2) → MK unwrap 3-layer hierarchy
  - KEK slot model supporting passphrase + recovery mnemonic slots simultaneously
  - KDF metadata (`KDFAlg`, `KDFParams`, `WrapAlg`, `Domain`) included in KEK slots for future algorithm transition
- **New**: BIP39-based recovery mnemonic (24-word)
  - `lango security recovery setup` — mnemonic generation + slot addition
  - `lango security recovery restore` — MK unwrap via mnemonic + new passphrase setup
- **New**: `internal/security/config_store.go` — unified `security_config` table access (eliminates bootstrap/session store duplication)
- **Modified (BREAKING internal)**: Bootstrap pipeline 7 → 10 phases
  - New phases: LoadEnvelopeFile, UnwrapOrCreateMK, MigrateEnvelope
  - Phase order reversed: envelope load → passphrase → MK unwrap → DB open
  - `phaseEnsureDataDir` applies `Options.LangoDir` first
- **Modified**: DB encryption key = HKDF(MK, "lango-db-encryption") (raw 32-byte key, `PRAGMA key = "x'<hex>'"`)
  - No DB rekey needed on passphrase change (MK is immutable)
  - One-time migration + `PRAGMA rekey` for legacy installs (WAL-safe `VACUUM INTO` backup)
- **Modified**: `lango security change-passphrase` new (envelope re-wrap only, no data re-encryption)
  - `migrate-passphrase` deprecated
- **Modified**: `lango security status` default behavior changed to passphrase-free (keyring/keyfile non-interactive, graceful degradation)
- **Modified**: `openDatabase()` signature adds `rawKey bool` (`PRAGMA key = "x'<hex>'"` vs `'<passphrase>'`)
- **Added**: Crash recovery — `PendingMigration` + `PendingRekey` flags absorb crashes during migration
- **CryptoProvider interface remains unchanged** — zero impact on all consumers (SecretsStore, ConfigStore, KMS providers, tools)

## Capabilities

### New Capabilities

- `master-key-envelope`: MK/KEK 3-layer key hierarchy. Envelope file storage, KEK slot model, passphrase/mnemonic slot, KDF metadata, crash recovery flags, MK wrap/unwrap contract
- `recovery-mnemonic`: BIP39 24-word mnemonic-based recovery. Generation, verification, slot addition, restore flow

### Modified Capabilities

- `passphrase-management`: Passphrase role transitions from data key to KEK. `migrate-passphrase` deprecated, `change-passphrase` new (envelope re-wrap)
- `passphrase-acquisition`: Recovery mnemonic selection option added when envelope exists. `AcquireNonInteractive()` new (keyring/keyfile only, no prompt)
- `bootstrap-lifecycle`: Phase count 7→10, order reversed (envelope first, DB later), 3 new phases, `Options.LangoDir` added
- `db-encryption`: DB key derivation = HKDF(MK), raw key mode (`"x'<hex>'"`), one-time `PRAGMA rekey` migration, WAL-safe backup
- `cli-security-status`: Passphrase-free default behavior, envelope section added, non-interactive mini-bootstrap, graceful degradation

> Note: `encrypted-config-profiles`, `cli-secrets-management`, `keyfile-shred`, `keyring-security-tiering` have no spec-level changes since the CryptoProvider interface remains unchanged (only internal implementation changes).

## Impact

### Code

- **New files (10)**: `internal/security/envelope.go`, `envelope_file.go`, `mnemonic.go`, `config_store.go`, `migrate_envelope.go` (+ tests), `internal/cli/security/change_passphrase.go`, `recovery.go`
- **Modified files (8)**: `internal/security/local_provider.go`, `errors.go`, `internal/bootstrap/bootstrap.go`, `phases.go`, `pipeline.go`, `internal/cli/security/status.go`, `migrate.go`, `internal/session/ent_store.go`
- **Unchanged**: `crypto.go` (interface), `secrets_store.go`, `key_registry.go`, `configstore/store.go`, `composite_provider.go`, all KMS providers, wallet, P2P

### Dependencies

- **New**: `github.com/tyler-smith/go-bip39` (BIP39 mnemonic library)
- **Existing**: `golang.org/x/crypto/pbkdf2`, `golang.org/x/crypto/hkdf`

### Data / Storage

- **New file**: `~/.lango/envelope.json` (0600 permissions, JSON encoded)
- **DB impact**: One-time migration re-encrypts all `secrets` and `config_profiles` + (SQLCipher only) `PRAGMA rekey`
- **No schema changes**: Existing ent entities (`Key`, `Secret`, `ConfigProfile`) remain as-is

### User-facing

- **New commands**: `lango security change-passphrase`, `lango security recovery setup`, `lango security recovery restore`
- **Deprecated**: `lango security migrate-passphrase`
- **Behavior change**: `lango security status` — default behavior without passphrase prompt, shows envelope info
- **Migration UX**: "Upgrading encryption format (one-time migration)..." message displayed on first boot of existing installs

### Residual Risk

- If passphrase is lost while `PendingMigration=true` or `PendingRekey=true`, MK can be unwrapped via mnemonic but the DB cannot be opened since it uses the legacy passphrase key. Recovery requires restoring from the `VACUUM INTO` backup (`lango.db.pre-migration`). This is the intersection of (migration crash) and (passphrase loss), making it an extremely rare scenario, but it is documented in docs and CLI warnings.
