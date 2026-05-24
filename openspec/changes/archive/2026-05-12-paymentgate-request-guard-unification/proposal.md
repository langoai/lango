## Why

The direct-payment gate still treats an empty transaction receipt id as a denied business outcome instead of an input-validation failure. That makes a caller mistake look like a policy verdict and weakens operator-facing clarity.

## What Changes

- Make `paymentgate.Service.EvaluateDirectPayment` reject an empty transaction receipt id with an actionable validation error.
- Update regression coverage for that request guard.
- Sync `production-readiness` coverage for the payment-gate request-id validation.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: the payment gate now treats an empty transaction receipt id as a validation error rather than a denied business outcome.

## Impact

- Affected code: `internal/paymentgate/service.go`
- Affected tests: `internal/paymentgate/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
