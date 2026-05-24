## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the subtype/count filter and alternate-sort upgrade for the dead-letter backlog list.

#### Scenario: Page describes subtype/count filters and sort modes
- **WHEN** a reader opens `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find `latest_status_subtype`, `manual_retry_count_min`, `manual_retry_count_max`, and `sort_by` described
- **AND** they SHALL find `manual_retry_count`, `latest_manual_replay_at`, and `latest_status_subtype` described
