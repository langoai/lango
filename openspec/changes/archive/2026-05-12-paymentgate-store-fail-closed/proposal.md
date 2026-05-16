## Why

The direct-payment gate service currently assumes its receipt store is always wired. If the store is missing, gate evaluation can panic instead of returning an actionable fail-closed error. That weakens the production safety of a policy-critical boundary.

## What Changes

- Make `paymentgate.Service` fail closed when its receipt store is unavailable.
- Add regression coverage for missing-store gate evaluation.
- Sync `production-readiness` coverage for the direct-payment gate dependency guard.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: the direct-payment gate now returns an actionable store-unavailable error instead of panicking when its receipt store is missing.

## Impact

- Affected code: `internal/paymentgate/service.go`
- Affected tests: `internal/paymentgate/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
