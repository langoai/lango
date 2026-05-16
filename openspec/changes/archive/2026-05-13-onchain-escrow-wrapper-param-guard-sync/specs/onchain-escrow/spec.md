## ADDED Requirements

### Requirement: On-chain escrow tools keep actionable wrapper parameter guards

On-chain escrow tools SHALL reject missing required wrapper inputs with actionable parameter errors before downstream escrow logic begins.

#### Scenario: On-chain escrow create and resolve reject missing required inputs
- **WHEN** `escrow_create` is invoked without `milestones`
- **OR** `escrow_resolve` is invoked without `sellerPercent`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not proceed into downstream escrow execution logic
