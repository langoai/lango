## ADDED Requirements

### Requirement: Escrow adjudication request-id guards stay consistent

The escrow adjudication service SHALL treat an empty transaction receipt id as a validation error instead of a denied business outcome.

#### Scenario: Missing escrow adjudication transaction receipt id fails closed
- **WHEN** `escrowadjudication.Service.Adjudicate` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable `transaction receipt id is required` error
- **AND** SHALL not synthesize a denied `missing_receipt` result
