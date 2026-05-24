## Why

The workbench now advertises follow-up staging during an in-flight starter turn, but the UX contract also needs to include the editing path: once the follow-up draft exists, editing keys should behave like composer intent and the docs should say so.

## What Changes

- Keep the running-state copy focused on typing the next prompt and interrupting with `Enter`.
- Preserve the editing-intent path that returns focus to the composer when editing keys are used on a staged follow-up.
- Update docs and specs to mention that editing remains direct in the running-state follow-up loop.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Running-state follow-up guidance now explicitly includes direct editing as part of the loop.
- `downstream-docs-sync`: Public docs now describe focus-return editing during staged follow-up input.

## Impact

- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`
