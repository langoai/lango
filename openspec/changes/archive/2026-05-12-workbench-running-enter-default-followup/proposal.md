## Why

The running-state workbench copy already tells the operator that `Enter` can interrupt-and-run from the empty running state, but the actual behavior still only seeded the default follow-up draft instead of queuing it. That left a real mismatch between product contract and runtime behavior, and it kept the next-turn loop one keypress longer than it needed to be.

## What Changes

- Make `Enter` in the empty running-state queue the default context-aware follow-up immediately.
- Cover that behavior with a workbench regression test.
- Sync the mission-workbench spec so the no-draft running-state Enter path is explicit.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: `Enter` from the empty running-state now queues the default follow-up instead of only seeding it.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`, `internal/cli/cockpit/pages/missioncontrol_workbench.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`
