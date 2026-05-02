## ADDED Requirements

### Requirement: Durable mirror preserves approval identity semantics
The RunLedger durable mirror for built-in teammate approval blocking SHALL preserve the same approval identity semantics defined by the capability layer.

#### Scenario: Durable approval identity matches runtime policy
- **WHEN** a built-in teammate approval-blocked transition is mirrored into RunLedger
- **THEN** the durable record SHALL preserve the stable logical `grant_request_id`
- **AND** repeated attempts for the same logical blocked request SHALL NOT require rotating that request ID
