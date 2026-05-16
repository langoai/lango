## Why

The workbench already let the operator replace a staged follow-up with another starter prompt, but the contract should say that the replacement is not cosmetic: it becomes the actual next turn that runs when `Enter` is pressed.

## What Changes

- Clarify in docs and spec that replacing a staged follow-up with `1/2/3` changes the next prompt that will be executed.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Running follow-up starter replacement is now documented as an executable next-turn replacement, not just a text swap.
- `downstream-docs-sync`: Public docs now describe that the replacement starter prompt is what actually runs next.

## Impact

- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`
