## Why

The workbench activity lane now keeps assistant reply summaries, but those summaries can still carry raw multi-line or overly long text into the timeline. That weakens scanability and makes the result trail noisy right where the first-success loop is supposed to feel compact and controlled.

## What Changes

- Compact activity summaries at the shared buffer layer so they stay single-line and bounded in length.
- Cover multiline and long assistant summary cases with regression tests.
- Update public docs and the mission-workbench spec to state that the activity lane keeps a short timeline summary rather than replaying the full raw reply.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Activity summaries are compacted into short single-line timeline entries.
- `downstream-docs-sync`: Public docs describe the activity lane as a short summary trail rather than a raw reply mirror.

## Impact

- Affected code: `internal/cli/cockpit/activity_buffer.go`, `internal/cli/workbench/model.go`
- Affected tests: `internal/cli/cockpit/activity_buffer_test.go`, `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
