## MODIFIED Requirements

### Requirement: Dead-letter browsing and status observation tools
The meta tools surface SHALL provide read-only visibility into dead-lettered post-adjudication execution.

#### Scenario: Dead-letter backlog tool supports dominant family
- **WHEN** `list_dead_lettered_post_adjudication_executions` is invoked
- **THEN** it SHALL accept `dominant_family`
- **AND** each row SHALL expose `dominant_family`
