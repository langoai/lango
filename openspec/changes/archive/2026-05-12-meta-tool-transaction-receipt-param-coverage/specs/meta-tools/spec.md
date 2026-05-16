## ADDED Requirements

### Requirement: Transaction-receipt-backed execution tools keep actionable missing-parameter errors

Meta tools that require `transaction_receipt_id` for settlement or escrow execution SHALL reject missing values at the wrapper layer with actionable parameter errors before invoking service logic.

#### Scenario: Settlement and escrow execution tools reject missing transaction receipt ids
- **WHEN** `execute_settlement`, `execute_partial_settlement`, or `execute_escrow_recommendation` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not invoke the underlying execution service
