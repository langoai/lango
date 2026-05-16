## Why

`cmd/lango/main.go` now has three copies of the same startup-notice rendering pattern for `chat`, `cockpit`, and `workbench`: banner, log path, and initializing line. The behavior is already tested, but the duplication makes future edits more error-prone.

## What Changes

- Introduce a shared helper for TUI startup notice rendering in `cmd/lango`
- Reuse it from chat, cockpit, and workbench startup paths
- Keep the existing output contract and regressions unchanged

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: top-level TUI startup notice rendering is implemented through a shared helper

## Impact

- Affected code: `cmd/lango/main.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No runtime behavior changes
