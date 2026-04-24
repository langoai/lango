## MODIFIED Requirements

### Requirement: Dead-letter browsing and status observation tools
The meta tools surface SHALL provide read-only visibility into dead-lettered post-adjudication execution.

#### Scenario: Dead-letter backlog tool supports subtype/count filters and sort modes
- **WHEN** `list_dead_lettered_post_adjudication_executions` is invoked
- **THEN** it SHALL accept `latest_status_subtype`, `manual_retry_count_min`, `manual_retry_count_max`, and `sort_by`
- **AND** each row SHALL expose `manual_retry_count`, `latest_manual_replay_at`, and `latest_status_subtype`
