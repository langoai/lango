## ADDED Requirements

### Requirement: Economy escrow wrapper guards stay actionable

Economy-layer escrow tools SHALL preserve actionable missing-parameter errors for all declared wrapper inputs.

#### Scenario: Economy escrow tools reject missing required inputs
- **WHEN** `economy_escrow_create`, `economy_escrow_milestone`, `economy_escrow_status`, `economy_escrow_release`, or `economy_escrow_dispute` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before downstream escrow execution begins
