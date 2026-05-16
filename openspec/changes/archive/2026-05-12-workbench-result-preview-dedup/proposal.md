## Why

The completed-turn workbench body now surfaces the latest assistant summary, but successful turns currently read as `Last result: Assistant reply: ...`. That doubles the label and adds noise right where the operator wants the shortest possible scan path.

## What Changes

- Strip the `Assistant reply:` prefix from the completed-turn body preview while keeping failure prefixes such as `Turn timeout:` intact.
- Add regressions for the deduplicated result preview.
- Sync docs/specs for the cleaner preview wording.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: completed-turn body previews now show cleaner success summaries without duplicating the assistant label.
- `downstream-docs-sync`: public workbench docs describe the compact result preview more precisely.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
