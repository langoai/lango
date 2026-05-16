## Why

`p2p_pay` already enforces `transaction_receipt_id` at the wrapper layer, but `payment_send` still declared the same input as required while deferring the missing-input path to the downstream payment gate. The wrapper also treated `to`, `amount`, and `purpose` through one generic message instead of the standard missing-parameter contract.

## What Changes

- Tighten `payment_send` so all required wrapper inputs use actionable missing-parameter errors.
- Add regression coverage for the required wrapper inputs.
- Sync payment docs, prompts, and specs to the updated `payment_send` contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `payment-tools`: `payment_send` now preserves actionable wrapper-level missing-parameter errors for all required inputs.
- `production-readiness`: wrapper-level request-guard coverage now includes `payment_send`.
- `downstream-docs-sync`: payment tool docs and prompts describe the required receipt-linked payment contract accurately.

## Impact

- Affected code: `internal/tools/payment/payment.go`
- Affected tests: `internal/tools/payment/tools_test.go`
- Affected docs: `docs/payments/usdc.md`, `prompts/TOOL_USAGE.md`
- Affected specs: `openspec/specs/payment-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
