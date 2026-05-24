## Why

The runtime now shows key-accurate confirm prompts for both `a` and `s`, but the public cockpit feature docs still say the first press always warns with `Press 'a' again...`. That leaves the docs behind the actual approval-confirmation behavior.

## What Changes

- Update cockpit feature docs to explain that the second press repeats the same pending action key (`a` or `s`).
- Update the downstream-docs-sync spec to require the same key-accurate confirmation wording.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: Public cockpit docs need key-accurate approval confirm wording.

## Impact

- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
