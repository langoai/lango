## ADDED Requirements

### Requirement: Remaining transaction-receipt-backed decision tools keep actionable missing-parameter errors

Transaction-receipt-backed decision and update tools SHALL reject missing values at the wrapper layer with actionable parameter errors before invoking service logic.

#### Scenario: Path selection, approval, settlement progression, and escrow adjudication reject missing transaction receipt ids
- **WHEN** `select_knowledge_exchange_path`, `approve_upfront_payment`, `apply_settlement_progression`, or `adjudicate_escrow_dispute` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not invoke the underlying service
