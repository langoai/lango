## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing page describes refined retry recovery UX
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find refined cockpit retry state wording described
- **AND** they SHALL find CLI retry precheck rejection described separately from retry-request failure
- **AND** they SHALL find CLI retry success described as retry-request acceptance rather than completed execution

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with refined retry recovery UX and list the remaining work as dead-letter CLI `any_match_family` filtering, polling / follow-up recovery UX, richer structured CLI retry results, and broader operator summaries.

#### Scenario: Track page includes refined retry recovery UX as landed work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find refined retry success/failure wording described as landed work
- **AND** they SHALL find CLI retry precheck/request-accepted/request-failed semantics described as landed work
