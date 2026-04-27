## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the consolidated operator surface
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the landed operator-surface consolidation for post-adjudication dead letters.

#### Scenario: Page describes consolidated CLI and cockpit behavior
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find CLI `--any-match-family` filtering described
- **AND** they SHALL find grouped CLI and cockpit dispatch-family summaries described
- **AND** they SHALL find configurable top-N plus trend / time-window summary behavior described
- **AND** they SHALL find richer CLI and cockpit retry follow-up UX described

### Requirement: P2P knowledge exchange track reflects the completed operator-surface consolidation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the operator-surface consolidation work as landed and narrow the remaining backlog accordingly.

#### Scenario: Track page removes already-landed operator-surface gaps from the remaining work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find dead-letter CLI any-match-family filtering described as landed work
- **AND** they SHALL find grouped dispatch-family summaries described as landed work
- **AND** they SHALL find richer top-N / trend / time-window summaries described as landed work
- **AND** they SHALL find richer retry follow-up UX described as landed work
- **AND** they SHALL find the remaining work narrowed to broader taxonomy, history, retry-policy, and policy-default follow-ons
