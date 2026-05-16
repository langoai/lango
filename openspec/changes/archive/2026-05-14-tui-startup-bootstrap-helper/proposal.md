## Why

`runChat`, `runCockpit`, and `runWorkbench` still repeat the same logging/bootstrap/startup-notice preparation sequence even after their stderr seams were aligned. That duplication raises maintenance cost and makes future changes easier to drift.

## What Changes

- Extract a shared helper for top-level TUI logging/bootstrap/startup-notice preparation in `cmd/lango`
- Reuse it from chat, cockpit, and workbench
- Preserve the existing startup notice contract and existing regressions

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: top-level TUI startup bootstrap is implemented through a shared helper

## Impact

- Affected code: `cmd/lango/main.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No runtime behavior changes
