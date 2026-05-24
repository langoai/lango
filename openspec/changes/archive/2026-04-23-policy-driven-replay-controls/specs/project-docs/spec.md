## ADDED Requirements

### Requirement: Architecture landing and track docs reference policy-driven replay controls
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed policy-driven replay controls slice.

#### Scenario: Landing page links policy-driven replay controls
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Policy-Driven Replay Controls page listed with the other architecture pages

#### Scenario: Track doc reflects landed policy-driven replay controls
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** policy-driven replay controls SHALL be described as landed slice work with richer policy classes, policy editing surfaces, per-transaction snapshots, and amount-tier replay controls still remaining
