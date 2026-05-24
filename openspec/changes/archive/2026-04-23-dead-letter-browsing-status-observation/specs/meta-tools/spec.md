## ADDED Requirements

### Requirement: Dead-letter browsing and status observation tools
The system SHALL expose read-only tools for listing the current dead-letter backlog and inspecting one post-adjudication execution transaction.

#### Scenario: Backlog and detail tools are available
- **WHEN** the meta tools are built with a receipts store
- **THEN** both `list_dead_lettered_post_adjudication_executions` and `get_post_adjudication_execution_status` SHALL be available
