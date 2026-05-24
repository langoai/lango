## Why

The payment service still assumes core dependencies are always wired. A missing wallet provider, spending limiter, or transaction store can still crash user-facing paths like `send`, `balance`, `info`, or service-backed history/audit flows. That is incompatible with a production-safe CLI.

## What Changes

- Make the payment service fail closed when `wallet`, `limiter`, or `store` dependencies are missing.
- Add regression coverage for user-facing and service-facing paths that currently rely on those dependencies.
- Sync `production-readiness` coverage for payment core dependency wiring failures.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: payment core dependency wiring failures now return actionable errors instead of panicking.

## Impact

- Affected code: `internal/payment/service.go`
- Affected tests: `internal/payment/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
