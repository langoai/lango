## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Channel transcript rows stay compact and width-safe
- **WHEN** a `channel` transcript row renders long sender names or multiline remote message text on a narrow terminal
- **THEN** the row SHALL normalize those fields to a single line
- **AND** the rendered row SHALL clamp to the available transcript width instead of overflowing
