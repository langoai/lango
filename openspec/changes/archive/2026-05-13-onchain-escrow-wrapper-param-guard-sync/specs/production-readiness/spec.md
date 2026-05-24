## ADDED Requirements

### Requirement: On-chain escrow wrapper guards stay actionable

On-chain escrow tools SHALL preserve actionable missing-parameter errors for all declared wrapper inputs.

#### Scenario: On-chain escrow tools reject missing required inputs
- **WHEN** `escrow_create` or `escrow_resolve` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before downstream escrow execution begins
