## Why

The top-level `lango version` and `lango health` commands still print their success output via raw `fmt.Printf`/`fmt.Println`, which bypasses Cobra command output routing. That makes wrapper capture and command-level regression testing less consistent than the rest of the CLI.

## What Changes

- Route `lango version` success output through the Cobra command output stream
- Route `lango health` success output through the Cobra command output stream
- Add command-level regressions for both commands
- Update public docs and specs to describe the stream contract

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: top-level utility commands write success output to the command stream
- `cli-health-check`: successful health checks write `ok` to the command stream

## Impact

- Affected code: `cmd/lango/main.go`, `cmd/lango/main_test.go`
- Affected docs: `docs/cli/core.md`
- Affected specs: `openspec/specs/cli-reference/spec.md`, `openspec/specs/cli-health-check/spec.md`
- No feature expansion; this is stream-routing and testability hardening
