## Why

The repository ships a substantial payment and settlement runtime stack under `internal/finance`, `internal/paymentapproval`, `internal/paymentgate`, `internal/settlementprogression`, `internal/settlementexecution`, `internal/partialsettlementexecution`, `internal/escrowexecution`, `internal/disputehold`, `internal/escrowadjudication`, `internal/escrowrelease`, `internal/escrowrefund`, `internal/postadjudicationreplay`, and `internal/postadjudicationstatus`, but the public inventory docs omit those packages.

## What Changes

- add the current payment-settlement support packages to the README internal package tree
- add matching package rows to `docs/architecture/project-structure.md`
- add an executable guard that requires those package rows and their current responsibilities
- sync the main docs-only and test-coverage specs

## Impact

- more complete public inventory coverage for the payment and settlement runtime stack
- better discoverability of approval, settlement, escrow, and retry/status internals
- stronger regression protection against future omissions or vague wording
