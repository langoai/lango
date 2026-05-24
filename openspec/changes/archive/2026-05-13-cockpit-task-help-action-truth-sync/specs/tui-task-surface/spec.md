## ADDED Requirements

### Requirement: Tasks page help reflects the selected task's valid actions
The cockpit Tasks page SHALL expose cancel/retry help only when the currently selected task state supports that action.

#### Scenario: Running or pending task shows cancel help
- **WHEN** the selected task status is `running` or `pending`
- **THEN** the Tasks page help SHALL include the `c` cancel binding
- **AND** it SHALL NOT include retry-only help for that row

#### Scenario: Failed or cancelled task shows retry help
- **WHEN** the selected task status is `failed` or `cancelled`
- **THEN** the Tasks page help SHALL include the `r` retry binding
- **AND** it SHALL NOT include cancel-only help for that row

#### Scenario: Non-actionable task hides task action help
- **WHEN** the selected task status does not support cancel or retry
- **THEN** the Tasks page help SHALL omit both action bindings
