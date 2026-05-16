## Why

The receipts store already rejects missing `transaction_id`, `counterparty`, and `requested_scope` when opening a knowledge-exchange transaction, but the knowledge-runtime service does not directly lock that operator-facing contract in its own regressions.

## What Changes

- Add a regression test for missing canonical open inputs through `knowledgeruntime.Service.OpenTransaction`.
- Sync `production-readiness` coverage for the knowledge-runtime open-transaction request guard.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: knowledge-runtime open-transaction request validation now has explicit regression coverage.

## Impact

- Affected tests: `internal/knowledgeruntime/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
