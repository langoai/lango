## Why

The direct payment gate and `payment_send` tool now treat an empty `transaction_receipt_id` as an actionable validation error instead of a denied `missing_receipt` business outcome. Some payment specs and public docs still describe that path as a deny result, which leaves the documented contract behind the implemented one.

## What Changes

- Update the `payment-execution-gating` spec so empty request ids are modeled as validation errors, while unknown receipts remain deny outcomes.
- Update the `payment-tools` spec so `payment_send` documents the actionable validation error path for missing transaction receipt ids.
- Sync public payment/security docs with the implemented validation-vs-deny split.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `payment-execution-gating`: request validation and deny-reason semantics now match the implemented gate behavior.
- `payment-tools`: `payment_send` documents the actionable validation-error path for missing transaction receipt ids.
- `downstream-docs-sync`: payment/security docs describe the validation-vs-deny split accurately.

## Impact

- Affected docs: `docs/security/actual-payment-execution-gating.md`, `docs/payments/usdc.md`
- Affected specs: `openspec/specs/payment-execution-gating/spec.md`, `openspec/specs/payment-tools/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
