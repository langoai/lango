## ADDED Requirements

### Requirement: Economy escrow tools keep actionable wrapper parameter guards

Economy-layer escrow tools SHALL reject missing required wrapper inputs with actionable parameter errors before downstream escrow engine logic begins.

#### Scenario: Economy escrow tools reject missing required inputs
- **WHEN** `economy_escrow_create`, `economy_escrow_milestone`, `economy_escrow_status`, `economy_escrow_release`, or `economy_escrow_dispute` is invoked without one of its declared required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not proceed into downstream escrow engine operations
