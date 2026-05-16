## Why

The settlement and escrow meta tools already declare `transaction_receipt_id` as a required parameter, but their wrapper-level validation contract is only implied. The current protection lives in `toolparam.RequireString`, and without explicit regression coverage a future wrapper edit could start leaking weaker or inconsistent errors before the underlying services are reached.

## What Changes

- Add regression tests for `execute_settlement`, `execute_partial_settlement`, and `execute_escrow_recommendation` when `transaction_receipt_id` is omitted.
- Update the meta-tools capability spec to require actionable wrapper-level parameter validation for those tools.
- Add production-readiness coverage so the wrapper-level request guard remains explicit.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `meta-tools`: transaction-receipt-backed execution tools preserve actionable missing-parameter errors before service execution.
- `production-readiness`: meta tool wrapper request guards are explicitly covered for settlement/escrow execution tools.

## Impact

- Affected tests: `internal/app/tools_meta_settlementexecution_test.go`, `internal/app/tools_meta_partialsettlementexecution_test.go`, `internal/app/tools_meta_escrowexecution_test.go`
- Affected specs: `openspec/specs/meta-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
