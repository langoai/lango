## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the cross-submission lifecycle grouping upgrade for the dead-letter backlog list.

#### Scenario: Page describes transaction-global lifecycle grouping
- **WHEN** a reader opens `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find `transaction_global_total_retry_count`, `transaction_global_any_match_families`, and their filters described
