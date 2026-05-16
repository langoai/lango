## ADDED Requirements

### Requirement: Settlement progression request guards stay actionable

The settlement progression service SHALL preserve actionable validation errors for missing request/store prerequisites.

#### Scenario: Missing settlement progression transaction receipt id fails closed
- **WHEN** `settlementprogression.Service.ApplyReleaseOutcome` runs with an empty `transaction_receipt_id`
- **THEN** the call SHALL return `ErrInvalidApplyReleaseOutcomeRequest`
- **AND** SHALL preserve the actionable `transaction_receipt_id is required` cause

#### Scenario: Missing settlement progression receipt store fails closed
- **WHEN** `settlementprogression.Service.ApplyReleaseOutcome` runs without a configured receipt store
- **THEN** the call SHALL return `ErrInvalidApplyReleaseOutcomeRequest`
- **AND** SHALL preserve the actionable `receipt store is required` cause
