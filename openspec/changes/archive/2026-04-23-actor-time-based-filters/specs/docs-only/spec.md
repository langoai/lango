## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the actor/time-based filter upgrade for the dead-letter backlog list.

#### Scenario: Page describes actor/time-based filters
- **WHEN** a reader opens `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find `manual_replay_actor`, `dead_lettered_after`, and `dead_lettered_before` described
- **AND** they SHALL find `latest_dead_lettered_at` and `latest_manual_replay_actor` described
