## ADDED Requirements

### Requirement: Canonical open and approval wrapper guards stay actionable

The canonical knowledge-open and upfront-payment-approval tools SHALL preserve actionable missing-parameter errors for all required wrapper inputs.

#### Scenario: Open transaction and upfront approval reject missing required inputs
- **WHEN** `open_knowledge_exchange_transaction` or `approve_upfront_payment` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins
