## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the dead-letter browsing / status observation slice, including the richer list filtering/pagination surface and the current implementation limits.

#### Scenario: Page describes richer filtering and detail hints
- **WHEN** a reader opens `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find the backlog filters and pagination described
- **AND** they SHALL find the per-transaction navigation hints described
