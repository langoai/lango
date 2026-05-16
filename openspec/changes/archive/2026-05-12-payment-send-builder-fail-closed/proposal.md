## Why

The payment service still relies on panic-prone transaction builder wiring. In practice, a missing `TxBuilder` or a builder with no RPC client does not fail with an actionable error; it crashes the send path and forces tests to recover from the panic. That is not production-safe, and it breaks the broader fail-closed contract we have been tightening across runtime services.

## What Changes

- Make `TxBuilder.BuildTransferTx` fail closed when the builder itself or its RPC client is unavailable.
- Update `payment.Service.Send` coverage so builder-unavailable paths return actionable `build transaction` errors instead of panicking.
- Sync `production-readiness` spec coverage for builder-wiring failures in the payment send path.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: payment send wiring failures now preserve actionable builder-unavailable causes instead of crashing the send path.

## Impact

- Affected code: `internal/payment/tx_builder.go`
- Affected tests: `internal/payment/tx_builder_test.go`, `internal/payment/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
