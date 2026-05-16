## Why

The cockpit feature docs already describe the Sessions page as a newest-first summary list that distinguishes unavailable, empty, and failed load states. The README cockpit shortcut table still mentions only that the page exists and degrades when the source is absent.

## What Changes

- Update the README cockpit shortcut table to mention the Sessions page's newest-first ordering and explicit empty/unavailable split.
- Update the downstream-docs-sync spec to require the same README row coverage.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: README cockpit page descriptions need the current Sessions behavior contract.

## Impact

- Affected docs/specs: `README.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
