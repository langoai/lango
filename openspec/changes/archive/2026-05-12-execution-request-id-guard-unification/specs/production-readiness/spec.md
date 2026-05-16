## ADDED Requirements

### Requirement: Settlement-cluster execution request-id guards stay consistent

Execution services in the settlement cluster SHALL treat an empty transaction receipt id as a validation error instead of a denied business outcome.

#### Scenario: Escrow/dispute execution services reject empty transaction receipt ids
- **WHEN** `escrowrelease.Service.Execute`, `escrowrefund.Service.Execute`, `disputehold.Service.Execute`, or `partialsettlementexecution.Service.Execute` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable `transaction receipt id is required` error
- **AND** SHALL not synthesize a denied `missing_receipt` execution result
