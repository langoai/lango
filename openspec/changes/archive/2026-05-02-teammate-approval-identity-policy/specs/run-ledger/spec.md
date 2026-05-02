## ADDED Requirements

### Requirement: Durable mirror preserves approval identity semantics
The RunLedger durable mirror for built-in teammate approval blocking SHALL preserve both the stable logical `grant_request_id` and the latest attempt metadata for that logical request.

#### Scenario: Durable snapshot reflects renewed attempt without rotating request ID
- **WHEN** a built-in teammate approval-blocked request is re-issued for the same run and tool while the latest active blocked cycle is still in progress
- **THEN** the durable mirror SHALL preserve the same logical `grant_request_id`
- **AND** the latest durable snapshot SHALL reflect the new attempt metadata

#### Scenario: Durable mirror preserves only the latest active blocked cycle
- **WHEN** a prior blocked cycle for a logical request has already been cleared by grant or denial
- **THEN** a later blocked cycle for the same logical request MAY reuse the same `grant_request_id`
- **AND** the durable mirror SHALL only preserve the latest active cycle metadata for that logical request in the snapshot
