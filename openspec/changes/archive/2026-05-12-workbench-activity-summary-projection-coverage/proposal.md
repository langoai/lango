## Why

The workbench activity summary compaction contract is now enforced at the shared buffer layer, but the projection path that feeds Mission Control snapshots did not have a direct regression proving that compacted summaries survive into the rendered activity surface.

## What Changes

- Add a projector-level regression test that verifies long multi-line activity summaries remain compact after projection.
- Document that the activity projection keeps compact timeline summaries intact.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Compact assistant activity summaries are now directly verified through the Mission Control projection path.

## Impact

- Affected code: `internal/cli/cockpit/missioncontrol_projector_test.go`
- Affected docs: `openspec/specs/mission-workbench-tui/spec.md`
