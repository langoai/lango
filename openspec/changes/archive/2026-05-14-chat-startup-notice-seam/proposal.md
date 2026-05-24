## Why

`lango chat` still prints its startup notice directly to `os.Stderr` even though cockpit and workbench now use seam-aware stderr paths. That leaves one interactive top-level entrypoint less testable and less consistent than the others.

## What Changes

- Add a seam-aware stderr writer for chat startup notices
- Add boot/logging/app-builder seams so the startup notice path can be regression-tested without running the full TUI
- Add a command-level regression proving the startup notice is capturable before app construction fails
- Update public docs and CLI reference spec to include chat in the startup-notice contract

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: chat startup notices use a seam-aware stderr path

## Impact

- Affected code: `cmd/lango/main.go`, `cmd/lango/main_test.go`
- Affected docs: `docs/cli/core.md`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No command-surface expansion; this is TUI startup testability hardening
