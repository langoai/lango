## ADDED Requirements

### Requirement: Escrow execution docs describe the validation-vs-rejection split

Public security docs for escrow execution SHALL describe empty transaction receipt ids as validation failures and reserve receipt-backed rejection conditions for post-validation execution state.

#### Scenario: Escrow execution docs distinguish validation failures from rejection states
- **WHEN** a user reads the escrow execution doc
- **THEN** they SHALL find that an empty `transaction_receipt_id` returns an actionable validation error
- **AND** they SHALL find that approval, settlement-hint, missing-input, and already-progressed checks apply only after request validation succeeds
