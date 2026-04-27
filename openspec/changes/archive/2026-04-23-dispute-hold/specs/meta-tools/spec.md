## ADDED Requirements

### Requirement: Dispute hold meta tool
The system SHALL expose a receipts-backed meta tool for recording dispute hold evidence on funded dispute-ready escrow.

#### Scenario: Dispute hold tool available
- **WHEN** the meta tools are built with a receipts store and dispute hold runtime
- **THEN** `hold_escrow_for_dispute` SHALL be available
