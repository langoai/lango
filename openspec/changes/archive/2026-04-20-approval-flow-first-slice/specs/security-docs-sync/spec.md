## ADDED Requirements

### Requirement: Approval-flow operator docs
The security documentation set SHALL include an approval-flow document that describes first-slice artifact release approval, decision states, and audit-backed approval receipts.

#### Scenario: Approval-flow doc available
- **WHEN** a user reads the security documentation
- **THEN** they SHALL find a dedicated approval-flow document describing the first-slice approval model and its current limits

#### Scenario: Approval-flow docs linked from security index
- **WHEN** a user reads `docs/security/index.md`
- **THEN** they SHALL find a quick link to the approval-flow document
