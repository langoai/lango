## Context

Lango stores application configuration (API keys, bot tokens, OAuth secrets) in `lango.json` as plaintext. AI agents have filesystem and exec tool access, meaning they can read or exfiltrate these secrets. The `LANGO_PASSPHRASE` environment variable approach is also vulnerable since env vars can be leaked through the exec tool or process inspection.

The current initialization flow is: `config.Load(lango.json)` → `app.New(cfg)` → inline passphrase resolution in `wiring.go`. Security components (crypto, keys, secrets) are initialized mid-flight with passphrase sourced from env vars or config fields.

## Goals / Non-Goals

**Goals:**
- Eliminate plaintext secrets on disk by encrypting all configuration at rest (AES-256-GCM)
- Provide a unified bootstrap sequence that handles DB, crypto, and config loading before the app layer
- Support multiple configuration profiles for different environments
- Block agent access to the `~/.lango/` data directory
- Provide a migration path from existing `lango.json` files
- Support non-interactive environments (CI/Docker) via keyfile

**Non-Goals:**
- Full database encryption (SQLCipher) — only config blobs are encrypted, not session data
- Remote key management service integration (existing RPC/enclave providers are unchanged)
- GUI/web-based config editor
- Automatic `lango.json` deletion after migration (user decides when to remove)

## Decisions

### 1. Passphrase acquisition via priority chain (not env var)

**Decision**: Replace `LANGO_PASSPHRASE` env var with a structured acquisition chain: keyfile → interactive terminal → stdin pipe.

**Rationale**: Environment variables are readable by child processes and visible in `/proc`. A keyfile with 0600 permissions provides equivalent non-interactive support while being invisible to the exec tool's environment. Interactive prompt remains the most secure option for development.

**Alternatives considered**:
- Keep env var as fallback → rejected because it leaves the same attack surface
- OS keychain integration → rejected as too platform-specific and complex for a Go CLI tool

### 2. Single encrypted blob per profile (not per-field encryption)

**Decision**: Serialize entire `config.Config` to JSON, encrypt the whole blob, store as one `encrypted_data` column per profile.

**Rationale**: Simpler implementation, avoids partial decryption complexity, and the entire config is needed at startup anyway. Per-field encryption would add complexity with no practical benefit since all fields are needed simultaneously.

**Alternatives considered**:
- Per-field encryption with separate keys → rejected for unnecessary complexity
- Encrypt only sensitive fields, leave structure visible → rejected because field names reveal system capabilities

### 3. Bootstrap package as the single entry point

**Decision**: Create `internal/bootstrap/` that owns the entire startup sequence: DB open → passphrase → crypto → config profile → return `Result` struct.

**Rationale**: Eliminates scattered passphrase resolution across `wiring.go`, `crypto_init.go`, and env var checks. The `Result` struct carries all initialized components, preventing re-initialization and double DB opens.

**Alternatives considered**:
- Keep initialization distributed across `wiring.go` → rejected because it perpetuates the fragmented passphrase handling
- Make bootstrap a method on App → rejected because DB/crypto must be ready before App exists

### 4. Ent schema for ConfigProfile (not raw SQL)

**Decision**: Use ent ORM schema for `config_profiles` table, consistent with existing `Secret`, `Key`, `Session` schemas.

**Rationale**: Leverages existing ent code generation, auto-migration, and type-safe queries. The `security_config` table (salt/checksum) remains as raw SQL for backward compatibility with the existing session store.

### 5. Shared ent.Client via NewEntStoreWithClient

**Decision**: Bootstrap opens the DB once and passes the `*ent.Client` to both `configstore.Store` and `session.EntStore` via a new `NewEntStoreWithClient()` constructor.

**Rationale**: Avoids double DB open, double schema migration, and potential locking issues. The bootstrap owns the client lifecycle; session store just borrows it.

## Risks / Trade-offs

- **[Risk] Passphrase forgotten** → Mitigation: Checksum verification gives clear error. User can re-create profiles from backup JSON via `config import`.
- **[Risk] Keyfile stolen from disk** → Mitigation: 0600 permissions, `~/.lango/` blocked from agent access. Same security posture as SSH keys.
- **[Risk] Migration breaks existing deployments** → Mitigation: `lango.json` is auto-detected and migration offered on first run. Old file is never deleted automatically.
- **[Trade-off] All config encrypted as single blob** → Cannot query individual config fields without decrypting. Acceptable because config is always loaded entirely.
- **[Trade-off] No env var passphrase** → CI pipelines must switch to keyfile. Documented migration path provided.

## Migration Plan

1. Users run `lango serve` or any config command — bootstrap detects no profiles and finds `lango.json`
2. Auto-migration offered: encrypts existing config as "default" profile
3. Original `lango.json` preserved — user removes it manually when satisfied
4. CI/Docker: create `~/.lango/keyfile` with 0600 permissions, replace `LANGO_PASSPHRASE` env var
5. Rollback: `lango config export default > lango.json` restores plaintext config

## Open Questions

None — all decisions were resolved during implementation.
