## ADDED Requirements

### Requirement: Architecture landing and track docs reference escrow refund
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed escrow refund slice.

#### Scenario: Landing page links escrow refund
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Escrow Refund page listed with the other architecture pages

#### Scenario: Track doc reflects landed escrow refund
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** escrow refund SHALL be described as landed slice work with terminal-state design, dispute-linked refund branching, and release-after-refund safety work still remaining
