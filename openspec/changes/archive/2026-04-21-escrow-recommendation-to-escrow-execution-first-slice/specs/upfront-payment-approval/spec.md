## ADDED Requirements

### Requirement: Escrow recommendation binding
The upfront payment approval path SHALL bind escrow execution input onto the linked transaction receipt when the approved suggested mode is `escrow`.

#### Scenario: Escrow-approved request binds execution input
- **WHEN** an upfront payment approval outcome is approved with suggested mode `escrow`
- **THEN** the linked transaction receipt SHALL store escrow execution input and set escrow execution status to pending

#### Scenario: Non-escrow request leaves escrow execution state empty
- **WHEN** an upfront payment approval outcome does not recommend `escrow`
- **THEN** the linked transaction receipt SHALL not bind escrow execution input
