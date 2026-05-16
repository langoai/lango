## ADDED Requirements

### Requirement: Payment-settlement support package inventory guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped payment-settlement support packages or misdescribe their current responsibilities SHALL be enforced by an executable test.

#### Scenario: Payment-settlement package rows remain truthful
- **WHEN** the repository still ships `internal/finance`, `internal/paymentapproval`, `internal/paymentgate`, `internal/settlementprogression`, `internal/settlementexecution`, `internal/partialsettlementexecution`, `internal/escrowexecution`, `internal/disputehold`, `internal/escrowadjudication`, `internal/escrowrelease`, `internal/escrowrefund`, `internal/postadjudicationreplay`, and `internal/postadjudicationstatus`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those packages and their current responsibilities
- **AND** it SHALL fail if the docs fall back to stale or generic wording that omits approval evaluation, receipt gating, settlement progression, direct or partial settlement execution, escrow fund/hold/adjudication, release/refund execution, or post-adjudication retry/status projection
