## Why

Lango's configuration is stored as encrypted profiles in `~/.lango/lango.db` (AES-256-GCM), but legacy references to `lango.json` file-based configuration remain throughout the codebase and README. This creates confusion: users may attempt to create JSON config files directly instead of using `lango onboard` or `lango config` CLI. The cleanup unifies the configuration path around encrypted profiles and removes all legacy JSON guidance.

## What Changes

- **BREAKING**: Remove `--config` CLI flag from root command (no longer needed)
- **BREAKING**: Remove `Options.MigrationPath` from bootstrap (no automatic JSON migration on startup)
- Remove `candidateJSONPaths()` auto-search in bootstrap — first run creates a default profile directly
- Remove deprecated `config.Save()` function from `internal/config/loader.go`
- Add auto-deletion of source JSON file after `config import` for security
- Update `config export` to document passphrase verification requirement
- Rewrite `doctor` command to check encrypted profiles instead of JSON files
- Remove all `lango.json`, `${ENV_VAR}`, and `export API_KEY` references from README
- Add Docker headless configuration section documenting the import→delete pattern

## Capabilities

### New Capabilities

### Modified Capabilities

- `config-system`: Remove `Save()` function; JSON is import-only, not a primary storage format
- `bootstrap-lifecycle`: Remove `MigrationPath` option and `candidateJSONPaths()`; no automatic JSON migration
- `config-cli-commands`: `config import` now auto-deletes source file; `config export` documents passphrase requirement; `--config` root flag removed
- `cli-doctor`: Doctor checks encrypted profile existence instead of JSON file existence; Fix() guides to `lango onboard` instead of creating JSON
- `docker-deployment`: Add headless configuration documentation (import→delete pattern)

## Impact

- `cmd/lango/main.go`: `cfgFile` variable, `--config` flag, and all `MigrationPath` references removed
- `internal/bootstrap/bootstrap.go`: `Options.MigrationPath` field, `candidateJSONPaths()` function, and JSON migration branch in `handleNoProfile()` removed
- `internal/config/loader.go`: `Save()` function and `encoding/json` import removed
- `internal/configstore/migrate.go`: `os.Remove()` added after successful import
- `internal/cli/doctor/doctor.go`: JSON file search loop and `--config` flag removed; loads config via bootstrap
- `internal/cli/doctor/checks/config.go`: Rewritten to check encrypted profile instead of JSON file
- `README.md`: All JSON config examples, `${ENV_VAR}` references, and `export` commands removed
