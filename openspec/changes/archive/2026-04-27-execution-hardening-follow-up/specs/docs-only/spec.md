## MODIFIED Requirements

### Requirement: Dispute hold page describes the first funded dispute hold slice
The `docs/architecture/dispute-hold.md` page SHALL describe the current dispute hold slice with service-local serialization for concurrent attempts on the same transaction.

#### Scenario: Dispute hold docs mention serialized concurrent attempts
- **WHEN** a user reads `docs/architecture/dispute-hold.md`
- **THEN** they SHALL find concurrent hold attempts for the same transaction described as serialized inside the service boundary

### Requirement: Release-vs-refund adjudication page describes the first post-hold branching slice
The `docs/architecture/release-vs-refund-adjudication.md` page SHALL describe the current adjudication slice with service-local serialization for concurrent attempts on the same transaction.

#### Scenario: Adjudication docs mention serialized concurrent attempts
- **WHEN** a user reads `docs/architecture/release-vs-refund-adjudication.md`
- **THEN** they SHALL find concurrent adjudication attempts for the same transaction described as serialized inside the service boundary

### Requirement: Escrow refund page describes the first funded refund slice
The `docs/architecture/escrow-refund.md` page SHALL describe the current refund slice with service-local serialization for concurrent attempts on the same transaction.

#### Scenario: Escrow refund docs mention serialized concurrent attempts
- **WHEN** a user reads `docs/architecture/escrow-refund.md`
- **THEN** they SHALL find concurrent refund attempts for the same transaction described as serialized inside the service boundary

### Requirement: Retry / dead-letter handling page describes the first bounded retry slice
The `docs/architecture/retry-dead-letter-handling.md` page SHALL describe the current retry slice with canonical retry-key dedup and explicit panic-to-failed-task behavior.

#### Scenario: Retry docs mention dedup and panic behavior
- **WHEN** a user reads `docs/architecture/retry-dead-letter-handling.md`
- **THEN** they SHALL find canonical retry-key dedup across pending, running, and scheduled tasks described
- **AND** they SHALL find background-runner panics described as explicit task failures rather than orphaned running tasks
