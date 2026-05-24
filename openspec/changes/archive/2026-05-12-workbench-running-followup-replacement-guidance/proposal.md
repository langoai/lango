## Why

The workbench now lets the operator replace a staged follow-up with `1/2/3` while a starter turn is still running, but the public contract should say that explicitly.

## What Changes

- Update running-state docs to mention starter replacement while a follow-up draft is staged.
- Update the Mission Workbench spec to reflect that `1/2/3` remain active in the running follow-up state.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Running-state follow-up guidance now explicitly includes starter replacement as part of the loop.
- `downstream-docs-sync`: Public docs now describe `1/2/3` replacement while a follow-up draft is staged.

## Impact

- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`
