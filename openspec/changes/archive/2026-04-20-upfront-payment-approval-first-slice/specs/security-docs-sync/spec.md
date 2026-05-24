## ADDED Requirements

### Requirement: Upfront payment approval operator docs
The security documentation set SHALL include an upfront payment approval document that describes the first slice, including structured decision states, suggested payment modes, and transaction-level payment approval updates.

#### Scenario: Upfront payment approval doc available
- **WHEN** a user reads the security documentation
- **THEN** they SHALL find a dedicated upfront payment approval document describing the first slice and its current limits

#### Scenario: Upfront payment approval docs linked from index
- **WHEN** a user reads `docs/security/index.md`
- **THEN** they SHALL find a quick link to the upfront payment approval document
