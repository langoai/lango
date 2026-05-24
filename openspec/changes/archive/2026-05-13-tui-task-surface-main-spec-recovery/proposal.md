## Why

Several archived task-surface changes updated the same `Tasks page navigation` requirement, and the current main `tui-task-surface` spec has dropped multiple landed navigation/help scenarios. That leaves the authoritative spec weaker than the implemented and tested Tasks page behavior.

## What Changes

- Restore the missing Tasks navigation/help scenarios in the main `tui-task-surface` capability.
- Keep the change spec-only so the main spec once again matches the already-landed runtime, tests, and docs.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-task-surface`: Recover lost `Tasks page navigation` scenarios in the main capability spec.

## Impact

- Affected specs: `openspec/specs/tui-task-surface/spec.md`
- No runtime code changes
