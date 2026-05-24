## Why

The workbench quick-start flow already reduced seeding and submit friction, but after the starter was actually submitted, the empty-state surface could still look like a pre-submit workspace. That makes the operator wait on a turn while the UI still looks idle.

## What Changes

- Once a starter prompt is submitted and the turn is streaming, replace quick-start or seeded-start guidance with a running-state hint.
- Update the empty composer placeholder and footer hint to reflect the same in-flight state.
- Add a workbench regression that locks in the running-state copy.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Submitted starter prompts now move the empty-workbench guidance into an explicit running-state message until the turn completes.
- `downstream-docs-sync`: Public docs now mention the running-state hint after submit.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`, `internal/cli/cockpit/pages/missioncontrol_workbench.go`, `internal/cli/chat/chat.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
