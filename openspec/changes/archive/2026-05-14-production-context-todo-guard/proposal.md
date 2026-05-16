## Why

Production Go code under `cmd/` and `internal/` currently happens to be free of `context.TODO()`, but that is not enforced repository-wide. The only executable guard today is x402-specific, which leaves the rest of the codebase vulnerable to low-signal placeholder context creeping back in.

## What Changes

- Add an executable repository test that rejects `context.TODO()` in non-test Go files under `cmd/` and `internal/`
- Replace the remaining test-only `context.TODO()` calls in turntrace retention tests with `context.Background()`
- Record the broader production-code guard in OpenSpec

## Capabilities

### New Capabilities

### Modified Capabilities
- `production-readiness`: production Go files are guarded against `context.TODO()` reintroduction
- `test-coverage`: production context-placeholder hygiene is executable

## Impact

- Affected code: `internal/testutil/production_quality_guard_test.go`, `internal/turntrace/retention_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`, `openspec/specs/test-coverage/spec.md`
- No runtime behavior changes
