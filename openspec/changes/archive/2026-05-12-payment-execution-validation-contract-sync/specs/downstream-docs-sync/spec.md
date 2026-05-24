## ADDED Requirements

### Requirement: Payment execution docs describe the validation-vs-deny split

Public payment and security docs SHALL describe empty transaction receipt ids as validation failures and reserve `missing_receipt` for missing referenced receipt state after request validation succeeds.

#### Scenario: Payment execution docs distinguish validation failures from deny reasons
- **WHEN** a user reads the direct payment execution gate docs
- **THEN** those docs SHALL describe an empty `transaction_receipt_id` as an actionable validation error
- **AND** SHALL describe `missing_receipt` as a deny reason for missing referenced receipt state rather than missing input
