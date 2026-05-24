## ADDED Requirements

### Requirement: Escrow release meta tool
The system SHALL expose a receipts-backed meta tool for releasing funded escrow from canonical settlement progression state.

#### Scenario: Escrow release tool available
- **WHEN** the meta tools are built with a receipts store and escrow release runtime
- **THEN** `release_escrow_settlement` SHALL be available
