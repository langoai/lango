## MODIFIED Requirements

### Requirement: Dead-letter browsing and status observation tools
The meta tools surface SHALL provide read-only visibility into dead-lettered post-adjudication execution.

#### Scenario: Backlog tool supports filtering and pagination
- **WHEN** `list_dead_lettered_post_adjudication_executions` is invoked
- **THEN** it SHALL accept adjudication, retry-attempt range, text query, offset, and limit inputs
- **AND** it SHALL return entries plus count, total, offset, and limit metadata

#### Scenario: Detail tool returns navigation hints
- **WHEN** `get_post_adjudication_execution_status` succeeds
- **THEN** it SHALL return the current canonical snapshot and latest retry / dead-letter summary
- **AND** it SHALL return `is_dead_lettered`, `can_retry`, and `adjudication`
