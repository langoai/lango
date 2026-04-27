## ADDED Requirements

### Requirement: Operator replay tool
The system SHALL expose a receipts-backed tool for replaying dead-lettered post-adjudication execution through the existing background dispatch path.

#### Scenario: Replay requires dead-letter evidence
- **WHEN** replay is attempted without dead-letter evidence
- **THEN** the replay tool SHALL deny execution
