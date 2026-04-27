## MODIFIED Requirements

### Requirement: Dead-letter browsing and status observation tools
The meta tools surface SHALL provide read-only visibility into dead-lettered post-adjudication execution.

#### Scenario: Dead-letter backlog tool supports actor/time-based filters
- **WHEN** `list_dead_lettered_post_adjudication_executions` is invoked
- **THEN** it SHALL accept `manual_replay_actor`, `dead_lettered_after`, and `dead_lettered_before`
- **AND** it SHALL return backlog entries that include `latest_dead_lettered_at` and `latest_manual_replay_actor`
