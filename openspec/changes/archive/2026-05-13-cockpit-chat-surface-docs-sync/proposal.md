## Why

The cockpit Chat page already exposes a distinct operator surface with send/newline keys, slash commands, and inline approval interrupts, but the public cockpit feature docs mostly describe adjacent concepts like approvals and runtime visibility instead of the page itself. That leaves the main operator reference underselling one of the core cockpit surfaces.

## What Changes

- Add a dedicated Chat section to the cockpit feature docs.
- Describe the primary Chat key surface, slash-command surface, and inline approval-interrupt controls using the current runtime contract.
- Update the downstream-docs-sync spec to require the same Chat operator-surface coverage.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: Public cockpit docs need a dedicated Chat operator-surface description.

## Impact

- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
