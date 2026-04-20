## ADDED Requirements

### Requirement: Dispute-ready receipt operator docs
The security documentation set SHALL include a dispute-ready receipt document that describes the first lite receipt slice, including submission receipts, transaction receipts, current submission pointer, canonical state, and event trail.

#### Scenario: Dispute-ready receipt doc available
- **WHEN** a user reads the security documentation
- **THEN** they SHALL find a dedicated dispute-ready receipt document describing the first slice and its current limits

#### Scenario: Dispute-ready receipt docs linked from index
- **WHEN** a user reads `docs/security/index.md`
- **THEN** they SHALL find a quick link to the dispute-ready receipt document
