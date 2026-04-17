## Why

After de-vectorizing retrieval, the application still exposes direct SQLite ownership across bootstrap, app wiring, and many CLI commands. This makes it impossible to enforce a single storage trust boundary or prevent direct DB bypass paths.

## What Changes

- Introduce a single-binary, auto-managed storage broker subprocess.
- Move SQLite open, schema migration, and auxiliary table/index initialization into the broker.
- Replace direct `*sql.DB` / `*ent.Client` exposure with a storage facade returned from bootstrap.
- Remove direct DB access from app/CLI/status/doctor paths and enforce protected-path / FD inheritance rules.

## Capabilities

### New Capabilities
- `brokered-storage`: broker-owned SQLite access and coarse-grained storage RPC surface.

### Modified Capabilities
- `bootstrap-lifecycle`: bootstrap order changes to spawn/open the broker before config/profile load.
- `session-store`: session persistence moves behind the storage facade.
- `knowledge-store`: knowledge/search access moves behind the storage facade.
- `cli-security-status`: security status reads move from direct DB access to broker-backed diagnostics.
- `os-sandbox-core`: protected-path and subprocess inheritance rules expand to cover broker-owned resources.

## Impact

- Affected code: bootstrap, app wiring, session/knowledge/memory facades, CLI status/doctor/settings/config profile paths, sandbox policy.
- New runtime pattern: single binary distribution with internally managed long-lived broker subprocess.
