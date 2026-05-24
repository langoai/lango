## Why

The wrapper-level parameter coverage now pins the transaction-receipt and outcome guards for many operator tools, but the canonical knowledge-open and upfront-payment-approval entrypoints still rely on implicit `toolparam` behavior for their other required inputs. Those are core operator flows and should have the same explicit regression coverage.

## What Changes

- Add wrapper-level missing-parameter regressions for the canonical open-transaction inputs of `open_knowledge_exchange_transaction`.
- Add wrapper-level missing-parameter regressions for the required approval inputs of `approve_upfront_payment`.
- Extend meta-tools and production-readiness coverage for these required wrapper parameters.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `meta-tools`: canonical open-transaction and upfront-payment approval tools now have explicit wrapper-level missing-parameter coverage for their required inputs.
- `production-readiness`: wrapper-level request-guard coverage now includes canonical open/approval input validation.

## Impact

- Affected tests: `internal/app/tools_meta_knowledgeruntime_test.go`, `internal/app/tools_meta_paymentapproval_test.go`
- Affected specs: `openspec/specs/meta-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
