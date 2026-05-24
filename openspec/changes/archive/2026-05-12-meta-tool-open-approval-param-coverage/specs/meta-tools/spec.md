## ADDED Requirements

### Requirement: Canonical open and approval tools keep actionable missing-parameter errors

The canonical knowledge-open and upfront-payment-approval tools SHALL reject missing required inputs at the wrapper layer with actionable parameter errors before invoking service logic.

#### Scenario: Open transaction and upfront approval reject missing required inputs
- **WHEN** `open_knowledge_exchange_transaction` is invoked without one of `transaction_id`, `counterparty`, `requested_scope`, `price_context`, or `trust_context`
- **OR** `approve_upfront_payment` is invoked without one of `transaction_receipt_id`, `submission_receipt_id`, `amount`, `trust_score`, `user_max_prepay`, or `remaining_budget`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not invoke the underlying service
