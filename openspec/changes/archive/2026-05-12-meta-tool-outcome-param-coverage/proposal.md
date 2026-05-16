## Why

The transaction-receipt wrapper guard work now covers missing `transaction_receipt_id` broadly, but the two decision-oriented tools that also require `outcome` still rely on implicit `toolparam` behavior without direct regression coverage. That leaves a smaller but still operator-facing parameter contract unpinned.

## What Changes

- Add wrapper-level missing-`outcome` regressions for `apply_settlement_progression` and `adjudicate_escrow_dispute`.
- Extend meta-tools and production-readiness coverage so those decision tools explicitly preserve actionable missing-parameter errors for `outcome`.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `meta-tools`: decision-oriented transaction-receipt tools now have explicit wrapper-level coverage for missing `outcome`.
- `production-readiness`: wrapper-level request-guard coverage now includes the remaining required decision parameter for settlement progression and escrow adjudication.

## Impact

- Affected tests: `internal/app/tools_meta_settlementprogression_test.go`, `internal/app/tools_meta_escrowadjudication_test.go`
- Affected specs: `openspec/specs/meta-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
