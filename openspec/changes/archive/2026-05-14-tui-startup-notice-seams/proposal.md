## Why

`lango cockpit` and the default workbench entrypoint still print their startup notices through raw `os.Stderr`. That works interactively, but it keeps the startup banner/log-path/initializing hints outside seam-aware capture and makes regressions harder to lock down.

## What Changes

- Add seam-aware stderr writers for cockpit and workbench startup notices
- Add boot/logging/app-builder seams so those startup notice paths can be regression-tested without running the full TUI
- Add regressions that prove startup notice capture before app construction fails
- Update CLI reference docs/spec for the stderr-seam contract

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: cockpit and workbench startup notices use seam-aware stderr paths

## Impact

- Affected code: `cmd/lango/main.go`, `cmd/lango/main_test.go`
- Affected docs: `docs/cli/core.md`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No command-surface expansion; this is TUI startup testability hardening
