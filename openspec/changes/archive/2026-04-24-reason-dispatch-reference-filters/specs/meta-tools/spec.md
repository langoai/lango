## MODIFIED Requirements

### Requirement: Dead-letter browsing and status observation tools
The meta tools surface SHALL provide read-only visibility into dead-lettered post-adjudication execution.

#### Scenario: Dead-letter backlog tool supports reason and dispatch filters
- **WHEN** `list_dead_lettered_post_adjudication_executions` is invoked
- **THEN** it SHALL accept `dead_letter_reason_query` and `latest_dispatch_reference`
- **AND** it SHALL keep the existing page response shape unchanged
