## MODIFIED Requirements

### Requirement: Dead-letter browsing and status observation tools
The meta tools surface SHALL provide read-only visibility into dead-lettered post-adjudication execution.

#### Scenario: Post-adjudication status tool exposes the latest matching background task
- **WHEN** `get_post_adjudication_execution_status` succeeds
- **THEN** it SHALL expose optional `latest_background_task`
- **AND** the object SHALL expose `task_id`, `status`, `attempt_count`, and `next_retry_at`
