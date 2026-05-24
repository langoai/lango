## ADDED Requirements

### Requirement: Broader meta tool wrapper request guards stay actionable

Additional transaction-receipt-backed operator tools SHALL preserve actionable missing-parameter errors at the wrapper layer.

#### Scenario: Dispute, escrow-release, refund, status, and replay tools reject missing transaction receipt ids
- **WHEN** `hold_escrow_for_dispute`, `release_escrow_settlement`, `refund_escrow_settlement`, `get_post_adjudication_execution_status`, or `retry_post_adjudication_execution` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins
