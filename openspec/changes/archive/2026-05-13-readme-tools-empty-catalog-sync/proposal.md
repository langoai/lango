## Why

The cockpit feature docs already describe the Tools page's two degraded states distinctly: missing catalog and configured-but-empty catalog. The README cockpit shortcut table still mentions only the unavailable-catalog case, so it underspecifies one of the current operator-visible empty states.

## What Changes

- Update the README cockpit shortcut table to mention the explicit no-categories state for the Tools page.
- Update the downstream-docs-sync spec to require the same README row coverage.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: README cockpit page descriptions need the Tools empty-catalog distinction.

## Impact

- Affected docs/specs: `README.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
