## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing page describes dead-letter CLI reason/dispatch filtering
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find `--dead-letter-reason-query` described
- **AND** they SHALL find `--latest-dispatch-reference` described

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with dead-letter CLI reason/dispatch filtering, and list the remaining work as richer dead-letter CLI filters beyond latest subtype / latest family / actor-time / reason-dispatch, richer CLI recovery UX, and broader operator summaries.

#### Scenario: Track page includes dead-letter CLI reason/dispatch filtering as landed work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find dead-letter CLI reason/dispatch filtering described as landed work
