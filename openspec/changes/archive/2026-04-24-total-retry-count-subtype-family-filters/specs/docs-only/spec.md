## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the total retry-count and subtype-family filter upgrade for the dead-letter backlog list.

#### Scenario: Page describes total retry-count and family filters
- **WHEN** a reader opens `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find `total_retry_count_min`, `total_retry_count_max`, and `latest_status_subtype_family` described
- **AND** they SHALL find `total_retry_count` and `latest_status_subtype_family` described
