## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the transaction-global dominant-family upgrade for the dead-letter backlog list.

#### Scenario: Page describes transaction-global dominant family
- **WHEN** a reader opens `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find `transaction_global_dominant_family` described
