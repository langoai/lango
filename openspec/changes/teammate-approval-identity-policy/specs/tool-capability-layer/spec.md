## ADDED Requirements

### Requirement: Teammate approval request identity policy
The teammate capability layer SHALL define a stable approval identity policy for repeated blocked requests against the same run and tool.

#### Scenario: Approval identity semantics are explicit
- **WHEN** a built-in teammate blocks on a dangerous in-scope tool
- **THEN** the runtime SHALL produce a documented approval request identity
- **AND** a later block for the same run and tool SHALL reuse the same logical `GrantRequestID`
- **AND** any retry or re-request distinction SHALL be represented through separate attempt metadata rather than by rotating the logical request ID
