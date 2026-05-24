## MODIFIED Requirements

### Requirement: Dead-letter browsing and status observation tools
The meta tools surface SHALL provide read-only visibility into dead-lettered post-adjudication execution.

#### Scenario: Dead-letter backlog tool exposes compact per-submission breakdown
- **WHEN** `list_dead_lettered_post_adjudication_executions` is invoked
- **THEN** each row SHALL expose `submission_breakdown`
- **AND** each breakdown item SHALL expose `submission_receipt_id`, `retry_count`, and `any_match_families`
