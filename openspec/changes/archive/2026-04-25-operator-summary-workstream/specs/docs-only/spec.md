## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing page describes the summary CLI surface
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find `lango status dead-letter-summary` described
- **AND** they SHALL find total dead letters, retryable count, adjudication buckets, and latest-family buckets described

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with the first dead-letter CLI summary surface and list the remaining work as dead-letter CLI `any_match_family` filtering, polling / follow-up recovery UX, richer structured CLI retry results, richer dead-letter summaries, and cockpit summary surfaces.

#### Scenario: Track page includes dead-letter CLI summary as landed work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find the first dead-letter CLI summary surface described as landed work
- **AND** they SHALL find richer dead-letter summaries and cockpit summary surfaces described as remaining work
