## Why

The workbench already supports a two-step `Enter` path: first `Enter` seeds the default starter prompt, second `Enter` submits it. But after seeding, the footer still looked like a generic chat surface instead of teaching the immediate next action.

## What Changes

- Show an explicit `Enter submits starter` footer hint once a starter prompt is armed.
- Keep the empty-state footer focused on `Enter default starter` / `1-3 starter prompts` only before a starter is seeded.
- Update tests and docs so the launch-to-submit path is discoverable from the UI copy itself.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Seeded starter prompts now advertise the immediate submit path through footer copy.
- `downstream-docs-sync`: Public docs now mention the two-step Enter path more explicitly.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
