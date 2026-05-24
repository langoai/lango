## ADDED Requirements

### Requirement: Payment gate dependency guards stay actionable

The direct-payment gate SHALL preserve actionable fail-closed behavior when its backing receipt store is unavailable.

#### Scenario: Missing payment-gate receipt store fails closed
- **WHEN** `paymentgate.Service.EvaluateDirectPayment` runs without a configured receipt store
- **THEN** the call SHALL return an error identifying the unavailable payment-gate receipt store
- **AND** SHALL not silently allow execution or panic
