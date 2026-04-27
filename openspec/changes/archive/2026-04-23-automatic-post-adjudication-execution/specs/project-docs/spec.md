## ADDED Requirements

### Requirement: Architecture landing and track docs reference automatic post-adjudication execution
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed automatic post-adjudication execution slice.

#### Scenario: Landing page links automatic post-adjudication execution
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Automatic Post-Adjudication Execution page listed with the other architecture pages

#### Scenario: Track doc reflects landed automatic post-adjudication execution
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** automatic post-adjudication execution SHALL be described as landed slice work with background execution, retry orchestration, automatic execution as policy default, and broader dispute engine integration still remaining
