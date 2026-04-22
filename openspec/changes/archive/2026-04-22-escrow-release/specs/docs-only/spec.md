## ADDED Requirements

### Requirement: Architecture landing and track docs reference escrow release
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed escrow release slice.

#### Scenario: Landing page links escrow release
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Escrow Release page listed with the other architecture pages

#### Scenario: Track doc reflects landed escrow release
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** escrow release SHALL be described as landed slice work with refund, dispute-linked escrow handling, and milestone-aware release still remaining
