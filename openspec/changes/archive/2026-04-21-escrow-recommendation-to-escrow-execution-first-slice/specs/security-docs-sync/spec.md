## ADDED Requirements

### Requirement: Escrow execution operator docs
The security documentation set SHALL include an escrow execution document that describes the first receipt-backed escrow execution slice, its operator entry points, and its current limits.

#### Scenario: Escrow execution doc available
- **WHEN** a user reads the security documentation
- **THEN** they SHALL find a dedicated escrow execution document describing the first `create + fund` execution slice and its current limits

#### Scenario: Escrow execution docs linked from index
- **WHEN** a user reads `docs/security/index.md`
- **THEN** they SHALL find a quick link to the escrow execution document
