## MODIFIED Requirements

### Requirement: Settlement progression architecture page describes the current progression slice
The `docs/architecture/settlement-progression.md` page SHALL describe the current transaction-level settlement progression slice for `knowledge exchange v1`, including renewed disagreement semantics and the current limits of the slice.

#### Scenario: Settlement progression page describes renewed disagreement semantics
- **WHEN** a user reads `docs/architecture/settlement-progression.md`
- **THEN** they SHALL find `dispute-ready` described as a public canonical path for renewed disagreement
- **AND** they SHALL find re-escalation from `partially-settled` described as preserving the canonical `partial_settlement_hint`
- **AND** they SHALL find `apply_settlement_progression` described as returning `dispute_lifecycle_status`

### Requirement: Dispute hold page describes canonical hold-active lifecycle state
The `docs/architecture/dispute-hold.md` page SHALL describe dispute hold as the canonical hold-entry point for funded dispute-ready escrow, including `dispute_lifecycle_status = hold-active`.

#### Scenario: Dispute hold page describes hold-active tool receipts
- **WHEN** a user reads `docs/architecture/dispute-hold.md`
- **THEN** they SHALL find hold success described as setting `dispute_lifecycle_status = hold-active`
- **AND** they SHALL find `hold_escrow_for_dispute` described as returning `dispute_lifecycle_status`

### Requirement: Release-vs-refund adjudication page describes lifecycle-preserving canonical adjudication
The `docs/architecture/release-vs-refund-adjudication.md` page SHALL describe canonical adjudication as atomically updating settlement progression while preserving dispute lifecycle state for downstream recovery.

#### Scenario: Adjudication page describes lifecycle-aware tool receipts
- **WHEN** a user reads `docs/architecture/release-vs-refund-adjudication.md`
- **THEN** they SHALL find `release` and `refund` described as atomic progression updates
- **AND** they SHALL find the active dispute lifecycle marker described as preserved through adjudication
- **AND** they SHALL find `adjudicate_escrow_dispute` described as returning `dispute_lifecycle_status`

### Requirement: Retry / dead-letter handling page describes canonical re-escalation
The `docs/architecture/retry-dead-letter-handling.md` page SHALL describe exhausted post-adjudication retries as preserving canonical adjudication while re-escalating settlement progression and dispute lifecycle state.

#### Scenario: Retry / dead-letter page describes re-escalation state
- **WHEN** a user reads `docs/architecture/retry-dead-letter-handling.md`
- **THEN** they SHALL find exhausted retries described as setting `settlement_progression_status = dispute-ready`
- **AND** they SHALL find exhausted retries described as setting `dispute_lifecycle_status = re-escalated`

### Requirement: P2P knowledge exchange track reflects dispute runtime completion
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the landed dispute runtime with canonical hold-active and re-escalated lifecycle behavior, richer settlement progression semantics, and dispute-linked tool receipts that expose `dispute_lifecycle_status`.

#### Scenario: Track page reflects the completed dispute runtime slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find `hold-active` and `re-escalated` described as landed canonical lifecycle states
- **AND** they SHALL find richer settlement progression semantics described as landed work
- **AND** they SHALL find dispute-linked tool receipts described as returning `dispute_lifecycle_status`
