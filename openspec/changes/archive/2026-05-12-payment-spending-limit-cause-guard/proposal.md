## Why

The payment service already wraps spending-limit failures, but the regression only checked for a broad `spending limit` prefix. That leaves room for the underlying limiter cause to disappear or for the failure to regress into an earlier validation path without a direct test catching it.

## What Changes

- Tighten `payment.Service.Send` tests so spending-limit failures preserve the underlying limiter cause.
- Assert that the limit failure does not collapse into an invalid-amount validation path.
- Sync the production-readiness spec to make that error shape explicit.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: spending-limit failures now have a direct regression for wrapped limiter-cause preservation.

## Impact

- Affected code: `internal/payment/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
