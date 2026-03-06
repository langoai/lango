## MODIFIED Requirements

### Requirement: initEconomy accepts payment components
The `initEconomy` function SHALL accept a `*paymentComponents` parameter in addition to existing parameters. This parameter SHALL be passed from `app.New()` where `initPayment` result is available.

#### Scenario: Payment components passed to initEconomy
- **WHEN** `app.New()` initializes the economy layer
- **THEN** the `paymentComponents` from `initPayment` is passed as the `pc` parameter to `initEconomy`

#### Scenario: Nil payment components handled gracefully
- **WHEN** `initEconomy` receives nil `paymentComponents`
- **THEN** escrow falls back to `noopSettler` and all other economy components initialize normally
