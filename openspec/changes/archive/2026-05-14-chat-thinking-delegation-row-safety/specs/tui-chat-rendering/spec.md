## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Thinking transcript rows stay compact and width-safe
- **WHEN** a `thinking` transcript row renders preview text on a narrow terminal
- **THEN** the row SHALL stay single-line
- **AND** the rendered row SHALL clamp to the available transcript width instead of overflowing

#### Scenario: Delegation transcript rows stay compact and width-safe
- **WHEN** a `delegation` transcript row renders long actor names or a multiline reason on a narrow terminal
- **THEN** the row SHALL stay single-line
- **AND** the rendered row SHALL clamp to the available transcript width instead of overflowing
