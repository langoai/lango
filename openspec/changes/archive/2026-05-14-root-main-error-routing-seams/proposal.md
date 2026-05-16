## Why

`cmd/lango` still handles broker-mode and top-level root-command failures through raw process stdio and a direct `os.Exit(1)` path. That makes the entrypoint harder to test and leaves root failure routing less consistent than the seam-aware utility commands around it.

## What Changes

- Refactor the main entrypoint through a `runMain()` helper that returns an exit code
- Route broker-mode and root-command failure messages through injected stderr
- Replace the direct `os.Exit(1)` root failure path with the existing exit seam
- Add regressions for broker-mode failure and root-command failure

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: root entrypoint failure paths are seam-aware and stderr-routed under test

## Impact

- Affected code: `cmd/lango/main.go`, `cmd/lango/main_test.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No command-surface expansion; this is entrypoint testability and error-routing hardening
