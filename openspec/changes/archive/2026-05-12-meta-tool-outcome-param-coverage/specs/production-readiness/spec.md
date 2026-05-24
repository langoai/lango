## ADDED Requirements

### Requirement: Decision-tool wrapper outcome guards stay actionable

Decision-oriented transaction-receipt tools SHALL preserve actionable missing-parameter errors for `outcome` at the wrapper layer.

#### Scenario: Settlement progression and escrow adjudication reject missing outcomes
- **WHEN** `apply_settlement_progression` or `adjudicate_escrow_dispute` is invoked without `outcome`
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins
