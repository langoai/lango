## ADDED Requirements

### Requirement: Mission service request guards stay actionable

The mission service SHALL preserve actionable validation errors for required mission lifecycle inputs.

#### Scenario: Missing accepted-proposal source kind fails closed
- **WHEN** `mission.Service.AcceptProposal` runs without a `source_kind`
- **THEN** the call SHALL return an actionable `accept proposal: source_kind is required` error

#### Scenario: Missing execution reference fails closed
- **WHEN** `mission.Service.AttachExecution` runs without an `execution_ref`
- **THEN** the call SHALL return an actionable `attach execution: execution_ref is required` error
