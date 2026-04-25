## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing page describes dead-letter CLI actor/time filtering
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find `--manual-replay-actor` described
- **AND** they SHALL find `--dead-lettered-after` described
- **AND** they SHALL find `--dead-lettered-before` described

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with dead-letter CLI actor/time filtering, and list the remaining work as richer dead-letter CLI filters beyond latest subtype / latest family / actor-time, richer CLI recovery UX, and broader operator summaries.

#### Scenario: Track page includes dead-letter CLI actor/time filtering as landed work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find dead-letter CLI actor/time filtering described as landed work
