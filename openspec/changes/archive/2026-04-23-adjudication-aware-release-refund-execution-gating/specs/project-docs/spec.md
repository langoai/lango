## ADDED Requirements

### Requirement: Architecture landing and track docs reference adjudication-aware release/refund gating
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed adjudication-aware release/refund execution gating slice.

#### Scenario: Landing page links adjudication-aware gating
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Adjudication-Aware Release/Refund Execution Gating page listed with the other architecture pages

#### Scenario: Track doc reflects landed adjudication-aware gating
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** adjudication-aware release/refund execution gating SHALL be described as landed slice work with automatic post-adjudication execution, keep-hold or re-escalation states, and broader dispute engine integration still remaining
