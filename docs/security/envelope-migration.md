---
title: Master Key Envelope Migration
---

# Master Key Envelope Migration

When upgrading from a legacy Lango installation (passphrase-direct encryption) to the envelope-based architecture, an automatic one-time migration converts all encrypted data to the new format.

## What Changes

| Aspect | Legacy | Envelope |
|--------|--------|----------|
| Data encryption key | Passphrase-derived (PBKDF2) | Random 32-byte Master Key (MK) |
| Passphrase role | Direct data encryption | KEK source (wraps/unwraps MK only) |
| Passphrase change | Re-encrypts all data | Re-wraps MK only (O(1)) |
| Recovery | None (passphrase lost = data lost) | BIP39 mnemonic slot |
| DB key (SQLCipher) | Passphrase string | HKDF(MK, "lango-db-encryption") raw hex |
| Key metadata | Salt + checksum in DB | `envelope.json` on filesystem |

## Migration Process

On first boot after upgrade, bootstrap detects the legacy format (salt/checksum in DB, no `envelope.json`) and runs the migration automatically.

### Plaintext DB

```
1. Generate random 32-byte MK
2. Create envelope with passphrase KEK slot
3. Set PendingMigration = true, write envelope.json
4. BEGIN TX
   - Re-encrypt all secrets: old key decrypt -> MK encrypt -> UPDATE
   - Re-encrypt all config_profiles: old key decrypt -> MK encrypt -> UPDATE
   - COUNT verification (row counts match before/after)
   COMMIT
5. Clear PendingMigration, write envelope.json
```

### SQLCipher (Encrypted DB)

```
1-4. Same as plaintext
5. Set PendingRekey = true (PendingMigration cleared)
6. PRAGMA wal_checkpoint(TRUNCATE)
7. VACUUM INTO 'lango.db.pre-migration'  (WAL-safe backup)
8. PRAGMA rekey = "x'<HKDF(MK)>'"        (DB key rotation)
9. Close + reopen with MK-derived key     (verification)
10. Clear PendingRekey, write envelope.json
11. Remove backup file
```

## Crash Recovery

The migration uses two flags (`PendingMigration`, `PendingRekey`) stored in `envelope.json` to ensure safe recovery from any crash point.

| Crash Point | DB State | Recovery |
|-------------|----------|----------|
| Before envelope written | Legacy | No change, normal legacy boot |
| During data re-encryption | Legacy key, partial MK data | Next boot: passphrase opens DB, migration retries |
| After data, before rekey | Legacy key, MK data | Next boot: passphrase opens DB, rekey retries |
| During rekey | Unknown | Next boot: try MK key first, fallback to passphrase |
| After rekey, before flag clear | MK key, MK data | Next boot: MK key opens DB, flag cleared |

## Backup and Restore

### Automatic Backup

For SQLCipher migrations, a backup is created at `~/.lango/lango.db.pre-migration` via `VACUUM INTO` (WAL-safe). This backup uses the legacy passphrase as the DB key.

### Manual Backup

Before upgrading, you can manually back up:

```bash
cp ~/.lango/lango.db ~/.lango/lango.db.backup
```

### Restore from Backup

If migration fails and the backup exists:

```bash
# Stop Lango
cp ~/.lango/lango.db.pre-migration ~/.lango/lango.db
rm ~/.lango/envelope.json
# Restart — Lango boots in legacy mode
```

## Known Limitations

### Residual Risk: Pending State + Passphrase Lost

If a crash occurs during migration (`PendingMigration` or `PendingRekey` is true) AND the passphrase is subsequently lost, the recovery mnemonic can unwrap the MK but the DB is still keyed with the legacy passphrase. In this scenario:

- **With backup**: Restore `lango.db.pre-migration` and use the mnemonic to recover
- **Without backup**: Data is unrecoverable

This requires the intersection of (migration crash) AND (passphrase loss), which is extremely unlikely. The `VACUUM INTO` backup mitigates this risk for SQLCipher users.

### Legacy Data Retained

After migration, the legacy `security_config` row (salt + checksum) remains in the DB as a downgrade safety artifact. It is not consulted during envelope-based bootstrap.

## Verification

After migration completes, verify with:

```bash
lango security status
```

Expected output includes:

```
Master Key Envelope:
  Version:          1
  KEK Slots:        1 (passphrase)
  Recovery Setup:   disabled
```

To set up recovery after migration:

```bash
lango security recovery setup
```
