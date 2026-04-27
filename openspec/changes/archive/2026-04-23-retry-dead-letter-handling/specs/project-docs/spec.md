## ADDED Requirements

### Requirement: Architecture landing and track docs reference retry / dead-letter handling
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed retry / dead-letter handling slice.

#### Scenario: Landing page links retry / dead-letter handling
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Retry / Dead-Letter Handling page listed with the other architecture pages

#### Scenario: Track doc reflects landed retry / dead-letter handling
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** retry / dead-letter handling SHALL be described as landed slice work with operator replay, generic async retry policy, dead-letter browsing, and policy-driven backoff tuning still remaining
