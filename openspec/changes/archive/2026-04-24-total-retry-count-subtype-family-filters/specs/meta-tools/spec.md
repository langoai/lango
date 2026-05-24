## MODIFIED Requirements

### Requirement: Dead-letter browsing and status observation tools
The meta tools surface SHALL provide read-only visibility into dead-lettered post-adjudication execution.

#### Scenario: Dead-letter backlog tool supports total retry-count and family filters
- **WHEN** `list_dead_lettered_post_adjudication_executions` is invoked
- **THEN** it SHALL accept `total_retry_count_min`, `total_retry_count_max`, and `latest_status_subtype_family`
- **AND** each row SHALL expose `total_retry_count` and `latest_status_subtype_family`
