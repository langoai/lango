## Why

The payment balance path still assumes complete runtime wiring. If the service is constructed without a transaction builder or RPC client, `Balance` can crash instead of returning an actionable error. That breaks the production goal of fail-closed operator-facing behavior.

## What Changes

- Make `payment.Service.Balance` fail closed when the balance path is missing a builder or RPC client.
- Add regressions that verify the high-level `get balance` path preserves the underlying wiring cause.
- Sync `production-readiness` coverage for balance wiring failures.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: payment balance wiring failures now return actionable errors instead of panicking.

## Impact

- Affected code: `internal/payment/service.go`
- Affected tests: `internal/payment/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
