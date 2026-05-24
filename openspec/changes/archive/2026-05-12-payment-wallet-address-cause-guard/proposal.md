## Why

The payment service already wraps wallet-address lookup failures, but the current regression only checks for the broad `get wallet address` prefix and `ErrorIs`. That leaves room for the underlying wallet cause to disappear or for the failure to blur into earlier validation branches.

## What Changes

- Tighten `payment.Service.Send` tests so wallet-address failures preserve the underlying wallet cause.
- Assert that this failure path does not collapse into validation or spending-limit errors.
- Sync the production-readiness spec to make that error shape explicit.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: wallet-address lookup failures now have a direct regression for wrapped cause preservation.

## Impact

- Affected code: `internal/payment/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
