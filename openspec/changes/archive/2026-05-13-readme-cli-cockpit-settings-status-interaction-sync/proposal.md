## Why

The public cockpit feature docs now explain that Settings uses an embedded inline help footer and that Status is a read-only auto-refreshing page, but the README and CLI overview still describe those pages only in broad functional terms. That leaves first-touch operator docs behind the actual interaction model.

## What Changes

- Update the README cockpit shortcut table to describe the Settings inline footer and the Status read-only auto-refresh behavior.
- Update the `lango cockpit` CLI overview in `docs/cli/core.md` to mention the same interaction model at a high level.
- Update the downstream-docs-sync spec to require this README/CLI coverage.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: README and CLI overview need the current Settings and Status interaction semantics.

## Impact

- Affected docs/specs: `README.md`, `docs/cli/core.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
