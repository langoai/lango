## ADDED Requirements

### Requirement: Additional transaction-receipt-backed operator tools keep actionable missing-parameter errors

Transaction-receipt-backed operator tools beyond the direct settlement executors SHALL reject missing values at the wrapper layer with actionable parameter errors before invoking service logic.

#### Scenario: Dispute, escrow-release, refund, status, and replay tools reject missing transaction receipt ids
- **WHEN** `hold_escrow_for_dispute`, `release_escrow_settlement`, `refund_escrow_settlement`, `get_post_adjudication_execution_status`, or `retry_post_adjudication_execution` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not invoke the underlying service
