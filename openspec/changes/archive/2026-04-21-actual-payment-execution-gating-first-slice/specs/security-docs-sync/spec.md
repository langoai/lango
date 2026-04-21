## ADDED Requirements

### Requirement: Actual payment execution gating operator docs
The security documentation set SHALL include an actual payment execution gating document that describes the first direct-payment gate slice, its allow/deny behavior, and its current limits.

#### Scenario: Payment execution gate doc available
- **WHEN** a user reads the security documentation
- **THEN** they SHALL find a dedicated payment execution gating document describing the first slice and its current limits

#### Scenario: Payment execution gate docs linked from index
- **WHEN** a user reads `docs/security/index.md`
- **THEN** they SHALL find a quick link to the payment execution gating document
