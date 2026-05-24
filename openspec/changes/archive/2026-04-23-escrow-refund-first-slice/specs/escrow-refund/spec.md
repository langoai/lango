## ADDED Requirements

### Requirement: Receipt-backed escrow refund execution
The system SHALL provide a small refund execution service for the first refund
slice using `transaction_receipt_id` as the only canonical input.

#### Scenario: Missing receipt is denied
- **WHEN** the service is called without a transaction receipt id or the
  receipt cannot be found
- **THEN** the result SHALL be denied with `missing_receipt`

#### Scenario: Current submission is required
- **WHEN** the transaction has no current submission receipt
- **THEN** the result SHALL be denied with `no_current_submission`

#### Scenario: Funded escrow is required
- **WHEN** the transaction escrow execution status is not `funded`
- **THEN** the result SHALL be denied with `escrow_not_funded`

#### Scenario: Review-needed progression is required
- **WHEN** the settlement progression status is not `review-needed`
- **THEN** the result SHALL be denied with `not_review_needed`

#### Scenario: Refund amount must resolve
- **WHEN** the refund amount cannot be resolved from canonical transaction
  context
- **THEN** the result SHALL be denied with `amount_unresolved`

#### Scenario: Refund execution succeeds
- **WHEN** the receipt is funded, review-needed, and the amount resolves
- **THEN** the runtime SHALL execute the refund
- **AND** the result SHALL return `refund-executed`
- **AND** the result SHALL include transaction id, submission id, settled
  progression status still set to `review-needed`, resolved amount, and runtime
  reference

#### Scenario: Runtime failure preserves review-needed status
- **WHEN** the runtime returns an error
- **THEN** the result SHALL return a failure shape
- **AND** the result SHALL preserve `review-needed` in the result progression
  status
