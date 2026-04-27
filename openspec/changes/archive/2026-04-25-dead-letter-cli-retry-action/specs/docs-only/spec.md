## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing page describes dead-letter CLI retry action
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find `lango status dead-letter retry <transaction-receipt-id>` described
- **AND** they SHALL find `--yes` described

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with a dead-letter CLI retry action, and list the remaining work as richer dead-letter CLI filters beyond latest subtype / latest family, richer CLI recovery UX, and broader operator summaries.

#### Scenario: Track page includes dead-letter CLI retry action as landed work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find the dead-letter CLI retry action described as landed work
