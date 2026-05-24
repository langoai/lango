## Why

We refactored both `cmd/lango` and `cmd/zkexport` so exit handling flows through explicit seams, but nothing prevents a future edit from reintroducing direct `os.Exit(...)` calls into production entrypoint code.

## What Changes

- Add an executable repository test that rejects direct `os.Exit` references in `cmd/` production code outside the explicit seam declaration lines
- Record the guard in CLI reference and test-coverage specs

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: cmd entrypoint exit routing is regression-guarded
- `test-coverage`: cmd entrypoint exit hygiene has an executable guard

## Impact

- Affected code: `internal/testutil/cmd_entrypoint_exit_guard_test.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`, `openspec/specs/test-coverage/spec.md`
- No runtime behavior changes
