## ADDED Requirements

### Requirement: Actual settlement execution meta tool
The system SHALL expose a receipts-backed meta tool for executing direct final settlement from canonical settlement progression state.

#### Scenario: Actual settlement execution tool available
- **WHEN** the meta tools are built with a receipts store and settlement runtime
- **THEN** `execute_settlement` SHALL be available
