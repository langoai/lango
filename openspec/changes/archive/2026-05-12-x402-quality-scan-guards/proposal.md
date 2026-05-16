## Why

The production-readiness spec already requires that the `x402` package stay free of `context.TODO()` calls and legacy `NewX402Client` references, but that requirement was only implicit. A future refactor could reintroduce either issue without a direct guardrail.

## What Changes

- Add `x402` package quality-guard tests that scan the package source for `context.TODO()`.
- Add a codebase-wide guard that rejects legacy `NewX402Client` references.
- Keep the guard lightweight and file-based so it does not depend on external tooling.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: The x402 no-dead-code / no-`context.TODO()` requirement is now enforced by executable tests instead of relying on manual review.

## Impact

- Affected code: `internal/x402/quality_guard_test.go`
- Affected tests: `internal/x402`
