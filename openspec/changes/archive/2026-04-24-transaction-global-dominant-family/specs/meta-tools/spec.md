## MODIFIED Requirements

### Requirement: Dead-letter browsing and status observation tools
The meta tools surface SHALL provide read-only visibility into dead-lettered post-adjudication execution.

#### Scenario: Dead-letter backlog tool supports transaction-global dominant family
- **WHEN** `list_dead_lettered_post_adjudication_executions` is invoked
- **THEN** it SHALL accept `transaction_global_dominant_family`
- **AND** each row SHALL expose `transaction_global_dominant_family`
