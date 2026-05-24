## ADDED Requirements

### Requirement: Mission and proposal service dependency guards stay actionable

Core mission/proposal services SHALL preserve actionable fail-closed errors when required backing dependencies are missing.

#### Scenario: Missing mission store fails closed across mission lifecycle entrypoints
- **WHEN** mission lifecycle methods such as `StartMission`, `AcceptProposal`, or `AttachExecution` run without a configured mission store
- **THEN** each call SHALL return an error that identifies the attempted mission action
- **AND** SHALL preserve the actionable `mission store is required` cause instead of panicking

#### Scenario: Missing proposal registry fails closed across proposal entrypoints
- **WHEN** proposal service methods such as `UpsertLearningSuggestion`, `Accept`, or `PruneExpired` run without a configured proposal registry
- **THEN** each call SHALL return an error that identifies the attempted proposal action
- **AND** SHALL preserve the actionable `proposal registry is required` cause instead of panicking

#### Scenario: Missing proposal preparer fails closed during learning-suggestion preparation
- **WHEN** `proposal.Service.UpsertLearningSuggestion` runs with no configured proposal preparer
- **THEN** the call SHALL return an actionable `proposal preparer is required` error instead of panicking
