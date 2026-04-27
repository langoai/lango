## 1. Execution serialization

- [x] 1.1 Add service-local per-transaction serialization to `internal/disputehold`
- [x] 1.2 Add service-local per-transaction serialization to `internal/escrowadjudication`
- [x] 1.3 Add service-local per-transaction serialization to `internal/escrowrefund`
- [x] 1.4 Add focused concurrency tests for those services

## 2. Failure-path coverage

- [x] 2.1 Add dispute hold success-record partial-commit coverage
- [x] 2.2 Add escrow refund success-record partial-commit coverage

## 3. Background dedup and safety

- [x] 3.1 Deduplicate canonical retry-key submissions while tasks are pending, running, or scheduled
- [x] 3.2 Add focused dedup tests
- [x] 3.3 Document panic-to-failed-task policy inline in the background manager

## 4. Settlement progression safety

- [x] 4.1 Make `escalationProgressionStatus` exhaustive for known progression states
- [x] 4.2 Add an unknown-status panic test

## 5. Docs / OpenSpec

- [x] 5.1 Update public architecture docs for the landed hardening behavior
- [x] 5.2 Update docs-only OpenSpec requirements
