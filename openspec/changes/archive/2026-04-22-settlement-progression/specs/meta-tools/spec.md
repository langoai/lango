## ADDED Requirements

### Requirement: Settlement progression meta tool
The system SHALL expose a receipts-backed meta tool for applying artifact release outcomes to transaction-level settlement progression state.

#### Scenario: Settlement progression tool available
- **WHEN** the meta tools are built with a receipts store
- **THEN** `apply_settlement_progression` SHALL be available
