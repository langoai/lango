## ADDED Requirements

### Requirement: Settlement execution docs describe the validation-vs-deny split

Public settlement execution docs SHALL describe empty transaction receipt ids as validation failures and reserve deny reasons for referenced receipt or settlement state after request validation succeeds.

#### Scenario: Settlement execution docs distinguish validation failures from deny reasons
- **WHEN** a user reads the actual settlement execution docs
- **THEN** those docs SHALL describe an empty `transaction_receipt_id` as an actionable validation error
- **AND** SHALL describe `missing_receipt` and the other deny reasons as post-validation execution outcomes
