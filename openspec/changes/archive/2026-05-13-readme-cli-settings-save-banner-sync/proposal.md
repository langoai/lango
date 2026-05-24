## Why

The cockpit feature docs now explain that Settings saves surface inline success and failure banners, but the README and CLI overview still describe the page only as an editor with inline help and degraded save-unavailable messaging. That leaves first-touch operator docs behind one of the page's visible feedback loops.

## What Changes

- Update the README cockpit shortcut table to mention inline save feedback on the Settings page.
- Update the `lango cockpit` CLI overview to mention that embedded saves report success or failure inline.
- Update the downstream-docs-sync spec to require this README/CLI coverage.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: README and CLI overview need Settings save-feedback coverage that matches the embedded editor.

## Impact

- Affected docs/specs: `README.md`, `docs/cli/core.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
