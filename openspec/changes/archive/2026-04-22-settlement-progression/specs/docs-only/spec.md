## ADDED Requirements

### Requirement: Architecture landing and track docs reference settlement progression
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed settlement progression slice.

#### Scenario: Landing page links settlement progression
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Settlement Progression page listed with the other architecture pages

#### Scenario: Track doc reflects landed settlement progression
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** settlement progression SHALL be described as landed slice work with follow-on execution and dispute work still remaining
