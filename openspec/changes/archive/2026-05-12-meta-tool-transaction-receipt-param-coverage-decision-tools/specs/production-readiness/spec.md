## ADDED Requirements

### Requirement: Remaining meta tool wrapper request guards stay actionable

The remaining transaction-receipt-backed decision and update tools SHALL preserve actionable missing-parameter errors at the wrapper layer.

#### Scenario: Path selection, approval, settlement progression, and escrow adjudication reject missing transaction receipt ids
- **WHEN** `select_knowledge_exchange_path`, `approve_upfront_payment`, `apply_settlement_progression`, or `adjudicate_escrow_dispute` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins
