## ADDED Requirements

### Requirement: Architecture landing and track docs reference operator replay / manual retry
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed operator replay / manual retry slice.

#### Scenario: Landing page links operator replay / manual retry
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Operator Replay / Manual Retry page listed with the other architecture pages

#### Scenario: Track doc reflects landed operator replay / manual retry
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** operator replay / manual retry SHALL be described as landed slice work with dead-letter browsing UI, policy-driven replay controls, generic replay substrate design, and broader dispute engine integration still remaining
