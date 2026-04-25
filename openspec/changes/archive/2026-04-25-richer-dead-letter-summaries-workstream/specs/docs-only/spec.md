## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing page describes richer dead-letter summaries
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find top latest dead-letter reasons described on the summary CLI surface
- **AND** they SHALL find the top 5 limit described

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with richer dead-letter summaries and list the remaining work as dead-letter CLI `any_match_family` filtering, polling / follow-up recovery UX, richer structured CLI retry results, actor / dispatch summary breakdowns, grouped reason families or richer top-N/trend summaries, and cockpit summary surfaces.

#### Scenario: Track page includes richer dead-letter summaries as landed work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find top latest dead-letter reasons described as landed work
- **AND** they SHALL find actor / dispatch summary breakdowns and cockpit summary surfaces described as remaining work
