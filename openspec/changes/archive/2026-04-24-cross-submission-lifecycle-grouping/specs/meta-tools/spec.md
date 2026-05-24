## MODIFIED Requirements

### Requirement: Dead-letter browsing and status observation tools
The meta tools surface SHALL provide read-only visibility into dead-lettered post-adjudication execution.

#### Scenario: Dead-letter backlog tool supports transaction-global aggregation
- **WHEN** `list_dead_lettered_post_adjudication_executions` is invoked
- **THEN** it SHALL accept `transaction_global_total_retry_count_min`, `transaction_global_total_retry_count_max`, and `transaction_global_any_match_family`
- **AND** each row SHALL expose `transaction_global_total_retry_count` and `transaction_global_any_match_families`
