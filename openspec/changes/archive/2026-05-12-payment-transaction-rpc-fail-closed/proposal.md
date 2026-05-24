## Why

The payment service still assumes that its transaction RPC client is always wired once execution reaches submission and confirmation helpers. If `rpcClient` is missing, those paths can panic instead of returning actionable errors. That undermines the production fail-closed guarantees we have been tightening elsewhere in the payment service.

## What Changes

- Make `submitWithRetry` fail closed when the transaction RPC client is missing.
- Make `waitForConfirmation` fail closed when the receipt RPC client is missing.
- Add regression coverage and sync `production-readiness` for those helper-level wiring failures.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: payment transaction submission and confirmation helpers now return actionable RPC-unavailable errors instead of panicking.

## Impact

- Affected code: `internal/payment/service.go`
- Affected tests: `internal/payment/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
