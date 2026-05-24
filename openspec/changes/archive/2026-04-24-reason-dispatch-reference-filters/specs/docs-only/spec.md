## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the reason and dispatch-reference filter upgrade for the dead-letter backlog list.

#### Scenario: Page describes reason and dispatch filters
- **WHEN** a reader opens `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find `dead_letter_reason_query` and `latest_dispatch_reference` described
