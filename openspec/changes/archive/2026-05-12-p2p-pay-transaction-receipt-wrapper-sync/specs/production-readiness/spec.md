## ADDED Requirements

### Requirement: P2P payment wrapper request guards stay actionable

The `p2p_pay` tool SHALL preserve actionable missing-parameter errors for its required transaction receipt id before receipt-backed payment evaluation begins.

#### Scenario: Missing p2p-pay transaction receipt id fails at the wrapper
- **WHEN** `p2p_pay` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not defer that validation to the downstream payment gate
