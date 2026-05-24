## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing page describes grouped reason-family summaries
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find dead-letter CLI `by_reason_family` summary buckets described
- **AND** they SHALL find the CLI `By reason family` table section described
- **AND** they SHALL find the cockpit `reason families:` summary strip line described
- **AND** they SHALL find the initial reason-family taxonomy described as `retry-exhausted`, `policy-blocked`, `receipt-invalid`, `background-failed`, and `unknown`
- **AND** they SHALL find raw top latest dead-letter reasons described as still available alongside grouped reason-family summaries

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with grouped latest reason-family summaries in the CLI and cockpit surfaces while preserving raw top latest dead-letter reasons.

#### Scenario: Track page includes grouped reason-family summaries as landed work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find dead-letter CLI `by_reason_family` described as landed work
- **AND** they SHALL find the CLI `By reason family` table section described as landed work
- **AND** they SHALL find the cockpit `reason families:` summary strip line described as landed work
- **AND** they SHALL find the initial reason-family taxonomy described as `retry-exhausted`, `policy-blocked`, `receipt-invalid`, `background-failed`, and `unknown`
- **AND** they SHALL find raw top latest dead-letter reasons described as still available alongside grouped reason-family summaries
