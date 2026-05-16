## Why

The x402 quality guard already prevented legacy `NewX402Client` references from reappearing, but the original production-readiness scenario also implied that the legacy factory itself should not return in package source. That deserves a direct test too.

## What Changes

- Add a package-source scan that rejects any reintroduction of a `NewX402Client` function definition.
- Keep the existing repository-wide reference guard in place.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: x402 legacy-factory removal is now enforced both at the call-site level and at the package-definition level.

## Impact

- Affected tests: `internal/x402/quality_guard_test.go`
