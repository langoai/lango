## ADDED Requirements

### Requirement: Payment gate request-id guards stay consistent

The direct-payment gate SHALL treat an empty transaction receipt id as a validation error instead of a denied business outcome.

#### Scenario: Missing payment-gate transaction receipt id fails closed
- **WHEN** `paymentgate.Service.EvaluateDirectPayment` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable transaction-receipt-id-required error
- **AND** SHALL not synthesize a denied `missing_receipt` result
