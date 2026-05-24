## ADDED Requirements

### Requirement: Architecture landing and track docs reference release-vs-refund adjudication
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed release-vs-refund adjudication slice.

#### Scenario: Landing page links release-vs-refund adjudication
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Release vs Refund Adjudication page listed with the other architecture pages

#### Scenario: Track doc reflects landed release-vs-refund adjudication
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** release-vs-refund adjudication SHALL be described as landed slice work with adjudication-aware execution, keep-hold or re-escalation states, and broader dispute engine integration still remaining
