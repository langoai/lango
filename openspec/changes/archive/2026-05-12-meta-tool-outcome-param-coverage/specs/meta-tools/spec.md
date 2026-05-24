## ADDED Requirements

### Requirement: Decision-oriented meta tools keep actionable missing-outcome errors

Decision-oriented transaction-receipt tools SHALL reject missing `outcome` values at the wrapper layer with actionable parameter errors before invoking service logic.

#### Scenario: Settlement progression and escrow adjudication reject missing outcomes
- **WHEN** `apply_settlement_progression` or `adjudicate_escrow_dispute` is invoked without `outcome`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not invoke the underlying service
