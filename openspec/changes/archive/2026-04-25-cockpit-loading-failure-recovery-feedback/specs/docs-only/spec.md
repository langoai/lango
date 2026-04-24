## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing page describes retry loading/failure feedback
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find retry `running...` feedback described
- **AND** they SHALL find duplicate retry guarding while running described
- **AND** they SHALL find backend failure-string surfacing described

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with retry loading/failure feedback, and list the remaining work as richer cockpit filters beyond latest/any-match family and higher-level CLI surfaces.

#### Scenario: Track page includes retry loading/failure feedback as landed work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find retry loading/failure feedback described as landed work
