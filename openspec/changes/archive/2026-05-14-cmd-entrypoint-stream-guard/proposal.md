## Why

We hardened `cmd/lango` and `cmd/zkexport` around stream routing and seams, but those improvements are still enforced only by localized tests. A future edit could reintroduce raw print calls or direct standard-stream references in entrypoint code and quietly bypass the newer contracts.

## What Changes

- Add an executable repository test that rejects raw `fmt.Print*` calls in `cmd/` production code
- Reject direct `os.Stdin`/`os.Stdout`/`os.Stderr` references in `cmd/` production code outside the small set of explicit seam declarations
- Record the guard in CLI reference and test-coverage specs

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: cmd entrypoint stream routing is regression-guarded
- `test-coverage`: cmd entrypoint stream hygiene has an executable guard

## Impact

- Affected code: `internal/testutil/cmd_entrypoint_stream_guard_test.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`, `openspec/specs/test-coverage/spec.md`
- No runtime behavior changes
