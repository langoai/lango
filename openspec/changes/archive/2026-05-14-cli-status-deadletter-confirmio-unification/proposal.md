## Why

`lango status dead-letter retry` still owns a local yes/no parser even though the repository now has a shared confirmation helper with command-stream seams. That leaves another operator recovery path inconsistent with the rest of the CLI.

## What Changes

- Replace the inline dead-letter retry confirmation parser with the shared `prompt.ConfirmIO(...)` helper
- Preserve the current EOF-as-deny behavior so non-answers still abort cleanly
- Add a regression for the EOF/empty-input abort path
- Clarify the command-stream confirmation contract in docs/spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-status-dashboard`: dead-letter retry confirmation uses the shared helper and Cobra command streams

## Impact

- Affected code: `internal/cli/status/status.go`, `internal/cli/status/status_test.go`
- Affected docs/specs: `docs/cli/status.md`, `openspec/specs/cli-status-dashboard/spec.md`
- No feature expansion; this is a consistency and testability improvement
