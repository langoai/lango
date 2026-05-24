## ADDED Requirements

### Requirement: Meta tool wrapper request guards stay actionable

Transaction-receipt-backed settlement and escrow meta tools SHALL preserve actionable missing-parameter errors at the wrapper layer.

#### Scenario: Settlement and escrow meta tools reject missing transaction receipt ids
- **WHEN** `execute_settlement`, `execute_partial_settlement`, or `execute_escrow_recommendation` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins
