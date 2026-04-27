## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing page describes cockpit reset/clear shortcuts
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find `Ctrl+R` full filter reset described
- **AND** they SHALL find confirm-state clearing on reset described
- **AND** they SHALL find running-state no-op behavior described

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with cockpit reset/clear shortcuts, and list the remaining work as selection preservation and higher-level CLI surfaces.

#### Scenario: Track page includes cockpit reset/clear shortcuts as landed work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find cockpit reset/clear shortcuts described as landed work
