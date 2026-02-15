## Context

Lango's configuration has fully transitioned to encrypted SQLite profiles (`~/.lango/lango.db`) with passphrase-protected AES-256-GCM encryption. However, the codebase retains vestiges of the previous `lango.json` file-based approach: automatic JSON migration on startup, a `--config` CLI flag, a deprecated `config.Save()` function, and JSON-centric doctor checks. The README still instructs users to create `lango.json` directly.

This cleanup removes all JSON-as-primary-config paths and unifies around `lango onboard` (TUI) and `lango config` (CLI) as the sole configuration methods.

## Goals / Non-Goals

**Goals:**
- Remove all automatic JSON migration and `--config` flag from bootstrap and CLI
- Delete the deprecated `config.Save()` function
- Make `config import` auto-delete the source JSON file after successful import
- Rewrite doctor's ConfigCheck to verify encrypted profiles instead of JSON files
- Remove all `lango.json`, `${ENV_VAR}`, and `export API_KEY` instructions from README
- Document the headless (Docker/CI) import→delete pattern

**Non-Goals:**
- Removing `config.Load()` — still needed for `config import` path
- Changing the encrypted storage format or passphrase flow
- Adding new features beyond what the cleanup requires

## Decisions

1. **Source JSON auto-deletion on import**: `MigrateFromJSON()` deletes the source file with `os.Remove()` after successful import. Deletion failure logs a warning but does not fail the operation. This prevents plaintext secrets from lingering on disk.

2. **No `--config` flag**: The `--config` flag existed to specify a JSON migration source. With automatic migration removed, this flag serves no purpose. Users who need to import JSON use `lango config import` explicitly.

3. **Doctor checks encrypted profile, not JSON**: `ConfigCheck` now tests whether `cfg != nil` (already loaded by bootstrap) and validates it. If `cfg` is nil, it checks whether `lango.db` exists to give a specific error message. Fix action guides users to `lango onboard` instead of creating a JSON file.

4. **Passphrase verification for export**: Bootstrap already validates the passphrase via checksum comparison. Since `configExportCmd` calls `bootstrapForConfig()`, the passphrase is implicitly verified before any config can be exported. No additional prompt is needed — if bootstrap succeeds, the passphrase was correct.

## Risks / Trade-offs

- **Users with existing `lango.json` who haven't migrated**: They will no longer get automatic migration on startup. Mitigation: the default profile is created with `DefaultConfig()`, and they can manually run `lango config import lango.json` to migrate.

- **Source file deletion on import**: If the import succeeds but the user didn't intend to delete the source, data is lost. Mitigation: the `--help` text documents this behavior, and the command prints a clear message about deletion.

- **Docker Compose config mounting**: The existing `docker-compose.yml` mounts `lango.json` read-only. This pattern shifts to: COPY JSON into container, import, auto-delete. The docker-compose.yml file itself is not modified in this change — only the README documentation is updated.
