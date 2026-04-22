## ADDED Requirements

### Requirement: Architecture landing and track docs reference actual settlement execution
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed actual settlement execution slice.

#### Scenario: Landing page links actual settlement execution
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Actual Settlement Execution page listed with the other architecture pages

#### Scenario: Track doc reflects landed actual settlement execution
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** actual settlement execution SHALL be described as landed slice work with partial settlement, escrow completion, and dispute work still remaining
