## Why

The previous wrapper-guard passes covered the execution, dispute, release/refund, replay, and status tools, but several transaction-receipt-backed meta tools still relied on implicit `toolparam` behavior without direct regression coverage. These tools are still first-class operator entrypoints and should keep the same explicit missing-parameter contract.

## What Changes

- Add wrapper-level missing-parameter regressions for `select_knowledge_exchange_path`, `approve_upfront_payment`, `apply_settlement_progression`, and `adjudicate_escrow_dispute`.
- Extend meta-tools and production-readiness coverage to require actionable wrapper-level parameter errors for that remaining transaction-receipt-backed cluster.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `meta-tools`: the remaining transaction-receipt-backed decision/update tools now have explicit wrapper-level missing-parameter coverage.
- `production-readiness`: wrapper-level request-guard coverage now spans the broader transaction-receipt-backed operator cluster.

## Impact

- Affected tests: `internal/app/tools_meta_knowledgeruntime_test.go`, `internal/app/tools_meta_paymentapproval_test.go`, `internal/app/tools_meta_settlementprogression_test.go`, `internal/app/tools_meta_escrowadjudication_test.go`
- Affected specs: `openspec/specs/meta-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
