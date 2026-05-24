## ADDED Requirements

### Requirement: Architecture landing and track docs reference dispute hold
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed dispute hold slice.

#### Scenario: Landing page links dispute hold
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Dispute Hold page listed with the other architecture pages

#### Scenario: Track doc reflects landed dispute hold
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** dispute hold SHALL be described as landed slice work with release-vs-refund adjudication, explicit held-state design, and dispute engine integration still remaining
