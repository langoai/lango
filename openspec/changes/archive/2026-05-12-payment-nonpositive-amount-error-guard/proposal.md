## Why

The payment service already distinguishes parse failures from business-rule failures, but the zero/negative amount regression only checks for the positive-amount message. It does not prove that this path stays distinct from the generic `invalid amount` parse path.

## What Changes

- Tighten `payment.Service.Send` tests so non-positive amounts must keep the explicit `amount must be positive` business-rule error.
- Assert that those failures do not regress into the generic invalid-amount parse path.
- Sync the production-readiness spec to make that distinction explicit.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: non-positive payment amounts now have a direct regression that keeps their business-rule error distinct from parse failures.

## Impact

- Affected code: `internal/payment/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
