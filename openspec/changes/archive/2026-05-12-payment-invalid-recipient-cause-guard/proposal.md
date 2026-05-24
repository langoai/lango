## Why

The payment service already rejects malformed recipient addresses, but the regression only checked for a high-level `invalid recipient` substring. That leaves the underlying validation cause free to disappear even though it is part of the operator-facing debugging contract.

## What Changes

- Tighten `payment.Service.Send` tests so invalid recipient errors must preserve the wrapped validation cause and rejected input.
- Sync the production-readiness spec to make that error shape explicit.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: invalid recipient failures now have a direct regression for wrapped validation-cause preservation.

## Impact

- Affected code: `internal/payment/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
