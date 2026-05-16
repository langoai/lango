## Why

The workbench footer already distinguished starter seeding from starter submission, but the empty-state body still repeated the pre-seed quick-start copy after a starter had been loaded. That left a small contradiction in the first-run UX.

## What Changes

- Replace the ready-profile empty-state body copy with a submit-focused message once a starter prompt is armed.
- Keep the footer and body aligned so the next action is obvious after the first `Enter`.
- Update tests and docs to reflect the seeded-starter submit guidance.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Seeded starter prompts now switch the empty-state body from quick-start discovery to submit guidance.
- `downstream-docs-sync`: Public docs now mention that the workbench copy pivots to the submit step after seeding.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
