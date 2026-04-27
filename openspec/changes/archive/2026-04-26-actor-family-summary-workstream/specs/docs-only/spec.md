## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including grouped actor-family summaries on the CLI and cockpit operator surfaces while preserving raw top latest manual replay actors.

#### Scenario: Dead-letter browsing page describes grouped actor-family summaries
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find dead-letter CLI `by_actor_family` summary buckets described
- **AND** they SHALL find the CLI `By actor family` table section described
- **AND** they SHALL find the cockpit `actor families:` summary strip line described
- **AND** they SHALL find the initial actor-family taxonomy described as `operator`, `system`, `service`, and `unknown`
- **AND** they SHALL find raw top latest manual replay actors described as still available alongside grouped actor-family summaries

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with grouped latest actor-family summaries in the CLI and cockpit surfaces while preserving raw top latest manual replay actors.

#### Scenario: Track page includes grouped actor-family summaries as landed work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find dead-letter CLI `by_actor_family` described as landed work
- **AND** they SHALL find the CLI `By actor family` table section described as landed work
- **AND** they SHALL find the cockpit `actor families:` summary strip line described as landed work
- **AND** they SHALL find the initial actor-family taxonomy described as `operator`, `system`, `service`, and `unknown`
- **AND** they SHALL find raw top latest manual replay actors described as still available alongside grouped actor-family summaries
- **AND** the remaining work SHALL be described as grouped dispatch families, configurable actor-family taxonomy, and richer top-N / trend / time-window summaries
