## ADDED Requirements

### Requirement: Escrow refund meta tool
The system SHALL expose a receipts-backed meta tool for refunding funded escrow from canonical review-needed settlement state.

#### Scenario: Escrow refund tool available
- **WHEN** the meta tools are built with a receipts store and escrow refund runtime
- **THEN** `refund_escrow_settlement` SHALL be available
