## Why

`lango serve` still prints its startup banner and feature summary through raw stdout, even though other top-level utility commands now use Cobra command streams. That makes wrapper capture and command-level regressions less consistent than the rest of the CLI.

## What Changes

- Route `lango serve` startup banner through the Cobra command output stream
- Route `lango serve` startup summary through the Cobra command output stream
- Add command-level regression coverage using injected boot/app/logging/shutdown seams
- Update public docs and CLI reference spec with the startup stream contract

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: serve startup output writes through the Cobra command output stream

## Impact

- Affected code: `cmd/lango/main.go`, `cmd/lango/main_test.go`
- Affected docs: `docs/cli/core.md`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No command-surface expansion; this is startup-output routing and testability hardening
