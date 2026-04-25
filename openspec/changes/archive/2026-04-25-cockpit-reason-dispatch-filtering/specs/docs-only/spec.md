## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing page describes cockpit reason/dispatch filtering
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find cockpit `dead_letter_reason_query` filtering described
- **AND** they SHALL find cockpit `latest_dispatch_reference` filtering described

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with cockpit reason/dispatch filtering, and list the remaining work as reset/clear shortcuts, selection preservation, and higher-level CLI surfaces.

#### Scenario: Track page includes cockpit reason/dispatch filtering as landed work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find cockpit reason/dispatch filtering described as landed work
