## Why

CLI command implementations have largely been migrated to Cobra command streams and seam-aware writers, but nothing prevents future code from slipping back to raw `fmt.Print*` calls or direct `os.Stdout`/`os.Stderr` references. That would weaken captureability, testability, and wrapper safety.

## What Changes

- Add an executable repository test that rejects raw `fmt.Print*` calls in CLI production code
- Reject direct `os.Stdout`/`os.Stderr` references in CLI production code outside explicit seam files
- Record the guard in CLI stream-contract spec coverage

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: CLI production stream routing is regression-guarded
- `test-coverage`: CLI production stream hygiene has an executable guard

## Impact

- Affected code: `internal/testutil/cli_stream_quality_guard_test.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`, `openspec/specs/test-coverage/spec.md`
- No runtime behavior changes
