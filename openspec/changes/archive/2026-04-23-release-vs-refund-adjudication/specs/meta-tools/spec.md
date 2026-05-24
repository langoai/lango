## ADDED Requirements

### Requirement: Release vs refund adjudication meta tool
The system SHALL expose a receipts-backed meta tool for recording release-vs-refund branching on funded dispute-ready escrow with prior hold evidence.

#### Scenario: Adjudication tool available
- **WHEN** the meta tools are built with a receipts store
- **THEN** `adjudicate_escrow_dispute` SHALL be available
