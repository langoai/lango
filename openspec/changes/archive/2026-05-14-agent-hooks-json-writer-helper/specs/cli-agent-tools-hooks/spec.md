## ADDED Requirements
### Requirement: Agent hooks JSON snapshot stays uniformly pretty-printed
`lango agent hooks --json` SHALL continue to emit the same pretty-printed JSON snapshot through the shared CLI JSON writer path.

#### Scenario: Agent hooks JSON payload shape remains unchanged
- **WHEN** `lango agent hooks --json` renders the hook snapshot
- **THEN** it SHALL still emit valid pretty-printed JSON matching the existing field shapes
