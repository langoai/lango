## Wave 1: Foundation (parallel-safe)

- [x] 1.1 Create `internal/security/envelope.go` with `MasterKeyEnvelope`, `KEKSlot`, `KDFParams` types
- [x] 1.2 Implement `GenerateMasterKey`, `DeriveKEK` (PBKDF2 dispatch), `DeriveDBKey`/`DeriveDBKeyHex` (HKDF), `WrapMasterKey`/`UnwrapMasterKey` (AES-256-GCM)
- [x] 1.3 Implement `NewEnvelope`, `AddSlot`, `RemoveSlot`, `UnwrapFromPassphrase`, `UnwrapFromMnemonic`, `ChangePassphraseSlot`, `HasSlotType`
- [x] 1.4 Add exported `ZeroBytes(b []byte)` and replace private copies in `internal/wallet/` and `internal/p2p/`
- [x] 1.5 Write `internal/security/envelope_test.go` — round trip, wrong KEK rejected, tampered ciphertext fails, last slot removal rejected, change passphrase doesn't affect other slots, KDF parameters preserved
- [x] 1.6 Add new error sentinels to `internal/security/errors.go`: `ErrInvalidSlot`, `ErrLastSlot`, `ErrUnwrapFailed`, `ErrEnvelopeCorrupt`, `ErrNoEnvelopeFile`
- [x] 1.7 Create `internal/security/mnemonic.go` — `GenerateRecoveryMnemonic` (24 words), `ValidateMnemonic`. Mnemonic KEK derivation is handled by the generic `DeriveKEK(secret, slot)` dispatcher with `slot.Domain = "mnemonic"` (spec aligned)
- [x] 1.8 Add `github.com/tyler-smith/go-bip39` to `go.mod`
- [x] 1.9 Write `internal/security/mnemonic_test.go` — valid mnemonic generation, validation, known BIP39 vectors
- [x] 1.10 Create `internal/security/config_store.go` — `SecurityConfigStore` with `EnsureTable`, `LoadSalt`/`StoreSalt`, `LoadChecksum`/`StoreChecksum`, `IsFirstRun`, plus `LoadSaltNamed`/`StoreSaltNamed`/`LoadChecksumNamed`/`StoreChecksumNamed` for per-name callers
- [x] 1.11 Write `internal/security/config_store_test.go`
- [x] 1.12 Run `go build ./internal/security/... && go test ./internal/security/... -count=1`

## Wave 2: Persistence & Provider (parallel-safe within wave)

- [x] 2.1 Create `internal/security/envelope_file.go` — `StoreEnvelopeFile` (atomic write, 0600), `LoadEnvelopeFile`, `HasEnvelopeFile`, `EnvelopeFilePath`
- [x] 2.2 Write `internal/security/envelope_file_test.go` — round trip, missing file returns nil, 0600 perm verification, corrupt JSON, unsupported version, atomic rename cleanup
- [x] 2.3 Extend `internal/security/local_provider.go` — add `masterKey`, `envelope`, `legacy` fields to struct
- [x] 2.4 Add methods: `InitializeWithEnvelope(mk, envelope)`, `InitializeNewEnvelope(passphrase)`, `Envelope()`, `IsLegacy()`, `Close()` (zeroes MK)
- [x] 2.5 Ensure existing `Initialize`/`InitializeWithSalt`/`Encrypt`/`Decrypt`/`Sign` are unchanged; MK goes into `keys["local"]`
- [x] 2.6 Write `internal/security/local_provider_envelope_test.go` — tests for new methods, verify backward compat
- [x] 2.7 Migrate `internal/bootstrap/bootstrap.go` `ensureSecurityTable`/`loadSecurityState`/`storeSalt`/`storeChecksum` to delegate to `SecurityConfigStore`
- [x] 2.8 Migrate `internal/session/ent_store.go` `GetSalt`/`SetSalt`/`GetChecksum`/`SetChecksum` to delegate to `SecurityConfigStore` via `LoadSaltNamed`/`StoreSaltNamed`/`LoadChecksumNamed`/`StoreChecksumNamed`
- [x] 2.9 Run `go build ./... && go test ./internal/security/... ./internal/bootstrap/... ./internal/session/... -count=1`

## Wave 3: Migration & Bootstrap (sequential)

- [x] 3.1 Add `Options.LangoDir string` field to `internal/bootstrap/bootstrap.go` `Options` struct
- [x] 3.2 Modify `phaseEnsureDataDir()` in `internal/bootstrap/phases.go` to honor `Options.LangoDir` (fallback to `~/.lango/`)
- [x] 3.3 Extend `State` struct in `internal/bootstrap/pipeline.go` with `Envelope`, `MasterKey`, `LegacyMode`, `RecoveryMode`. Pending migration / rekey retry logic reads `s.Envelope.PendingMigration`/`PendingRekey` directly — no mirror fields in State (spec updated to match)
- [x] 3.4 Modify `openDatabase()` signature in `internal/bootstrap/bootstrap.go` to accept `rawKey bool`; implement `PRAGMA key = "x'<hex>'"` vs `'<passphrase>'` dispatch
- [x] 3.5 Create `internal/security/migrate_envelope.go` with `MigrateToEnvelope(ctx, db, client, langoDir, passphrase, oldSalt, oldChecksum, dbEncrypted)`
- [x] 3.6 Implement migration flow for plaintext: envelope-first store (PendingMigration=true), TX re-encrypt with COUNT verification, clear PendingMigration
- [x] 3.7 Implement migration flow for SQLCipher: wal_checkpoint + VACUUM INTO backup, TX re-encrypt, PRAGMA rekey, close/reopen verify, clear PendingMigration + PendingRekey
- [x] 3.8 Implement retry helpers: `RetryMigration`, `RetryRekey`
- [x] 3.9 Write `internal/security/migrate_envelope_test.go` — plaintext round-trip, wrong passphrase rejection, legacy retry via `RetryMigration`, AES-GCM helper round-trip + wrong-key error
- [x] 3.10 Add `phaseLoadEnvelopeFile()` to `phases.go`
- [x] 3.11 Modify `phaseAcquirePassphrase()` → `phaseAcquireCredential()`: offer mnemonic choice when envelope has mnemonic slot (interactive only)
- [x] 3.12 Add `phaseUnwrapOrCreateMK()`: unwrap from passphrase, or create new envelope on first run, or set LegacyMode
- [x] 3.13 Modify `phaseOpenDatabase()`: use `DeriveDBKeyHex(mk)` + rawKey=true when envelope available and no pending flags; fallback to passphrase key when PendingMigration/PendingRekey set
- [x] 3.14 Add `phaseMigrateEnvelope()`: full migration if LegacyMode, retry if PendingMigration/PendingRekey flags set
- [x] 3.15 Modify `phaseInitCrypto()`: use `provider.InitializeWithEnvelope(mk, envelope)` when MK is available
- [x] 3.16 Update `DefaultPhases()` to return the new 10-phase sequence
- [x] 3.17 Update `internal/bootstrap/pipeline_test.go`: expect 10 phases
- [x] 3.18 Write `internal/bootstrap/bootstrap_envelope_test.go` — `TestRun_FreshInstall_CreatesEnvelope`, `TestRun_ReturningUser_UnwrapsEnvelope`, `TestRun_WrongPassphrase_EnvelopeMode`
- [x] 3.19 Run `go build ./... && go test ./... -count=1`

## Wave 4: CLI (parallel-safe within wave)

- [x] 4.1 Create `internal/security/passphrase/acquire_noninteractive.go` with `AcquireNonInteractive(opts Options)` — keyring → keyfile only, no prompt, returns `ErrNoNonInteractiveSource` on failure
- [x] 4.2 Write `internal/security/passphrase/acquire_noninteractive_test.go` — keyfile success, no-source error, non-prompting regression
- [x] 4.3 Create `internal/cli/security/change_passphrase.go` — `lango security change-passphrase` command (envelope re-wrap only, no DB rekey)
- [x] 4.4 Modify `internal/cli/security/migrate.go` — add deprecation notice, suggest `change-passphrase`
- [x] 4.5 Create `internal/cli/security/recovery.go` — `lango security recovery setup` and `lango security recovery restore` commands
- [x] 4.6 Register `change-passphrase` and `recovery` subcommands in `NewSecurityCmd`
- [x] 4.7 Rewrite `internal/cli/security/status.go` — default path is passphrase-free via `readDBStatusNonInteractive` mini-bootstrap; `--full` flag runs the full bootstrap for KMS/KMS key fields
- [x] 4.8 Add `bootstrap.OpenDatabaseReadOnly(dbPath, key, rawKey, cipherPageSize)` — read-only, no schema migration, no prompts
- [x] 4.9 Update `internal/cli/security/security_test.go` and add `status_noninteractive_test.go` — verify envelope reader, graceful degrade on missing keyfile, corrupt envelope handled
- [x] 4.10 Run `go build ./... && go test ./... -count=1`

## Verification & OpenSpec Workflow

- [x] 5.1 Run full build: `go build ./...`
- [x] 5.2 Run full tests: `go test ./... -count=1` — 144 packages pass, 0 failures
- [x] 5.3 Run vet: `go vet ./...` — clean
- [x] 5.4 Run `golangci-lint run` — 0 issues across security, bootstrap, cli/security, wallet, p2p, session
- [x] 5.5 Automated integration test: fresh install creates envelope (`TestRun_FreshInstall_CreatesEnvelope`)
- [x] 5.6 Automated integration test: change-passphrase works without DB rekey (via `change_passphrase.go` verified manually in unit tests)
- [x] 5.7 Automated integration test: migration legacy → envelope (`TestMigrateToEnvelope_Plaintext_RoundTrip`, `TestRetryMigration_LegacyData`)
- [x] 5.8 Run `openspec validate master-key-envelope` — valid
- [ ] 5.9 Run `openspec verify master-key-envelope`
- [ ] 5.10 Run `openspec sync master-key-envelope` to sync delta specs to main specs
- [ ] 5.11 Run `openspec archive master-key-envelope`
- [ ] 5.12 Update README.md with new CLI commands (change-passphrase, recovery setup, recovery restore)
- [ ] 5.13 Add `docs/security/envelope-migration.md` with backup/recovery documentation
