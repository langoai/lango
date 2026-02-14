## Why

Sensitive information (API keys, bot tokens, OAuth secrets) stored in plaintext in `lango.json` can be read by AI agents through `fs_read`/`exec` tools, creating a Critical security gap. The root cause — plaintext config on disk — must be eliminated by encrypting all configuration at rest using a passphrase-derived key.

## What Changes

- **BREAKING**: Remove `lango.json` as the primary configuration source; all settings move to an AES-256-GCM encrypted SQLite database
- **BREAKING**: Remove `LANGO_PASSPHRASE` environment variable support; passphrase is now acquired via keyfile (`~/.lango/keyfile`) or interactive prompt
- **BREAKING**: Remove `SecurityConfig.Passphrase` field from config struct
- Introduce multi-profile configuration system (`default`, `staging`, `production`, etc.)
- Add `internal/passphrase/` package for structured passphrase acquisition chain (keyfile → interactive → stdin pipe)
- Add `internal/configstore/` package for encrypted profile CRUD operations
- Add `internal/bootstrap/` package for unified application startup (DB + crypto + config loading)
- Add `lango config` subcommands: `list`, `create`, `use`, `delete`, `import`, `export`, `validate`
- Block agent filesystem access to `~/.lango/` directory
- Add `LANGO_PASSPHRASE` to exec tool environment variable blacklist
- Register all config secrets with the output SecretScanner

## Capabilities

### New Capabilities
- `encrypted-config-profiles`: Encrypted storage and multi-profile management of application configuration in SQLite using AES-256-GCM
- `passphrase-acquisition`: Structured passphrase acquisition chain with keyfile, interactive terminal, and stdin pipe sources
- `bootstrap-lifecycle`: Unified application bootstrap sequence (database → passphrase → crypto → config profile)
- `config-cli-commands`: CLI subcommands for profile management (list, create, use, delete, import, export, validate)

### Modified Capabilities
- `config-system`: Configuration loading now goes through encrypted DB instead of plaintext JSON; `Save()` simplified; `SecurityConfig.Passphrase` removed
- `passphrase-management`: Passphrase source priority changed from env-var → config → interactive to keyfile → interactive → stdin
- `tool-filesystem`: Added `BlockedPaths` config field and validation to deny agent access to protected directories
- `tool-exec`: Added `LANGO_PASSPHRASE` to environment variable blacklist
- `output-secret-scanning`: Config-level secrets (provider keys, channel tokens) are now auto-registered with the scanner

## Impact

- **Packages created**: `internal/passphrase/`, `internal/configstore/`, `internal/bootstrap/`
- **Ent schema added**: `internal/ent/schema/config_profile.go` + generated ORM code
- **Breaking API change**: `app.New(*config.Config)` → `app.New(*bootstrap.Result)`
- **Session store**: Added `NewEntStoreWithClient()` for DB connection sharing
- **Migration path**: `lango config import lango.json` converts existing plaintext configs
- **CI/Docker**: Use `~/.lango/keyfile` (0600 permissions) instead of `LANGO_PASSPHRASE` env var
