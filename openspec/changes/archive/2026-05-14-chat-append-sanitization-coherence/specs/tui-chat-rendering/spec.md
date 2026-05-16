## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Display-only transcript fields are stored in sanitized form
- **WHEN** the chat model appends `system`, `status`, `channel`, or `delegation` transcript data that contains ANSI/OSC escape sequences
- **THEN** the stored transcript fields used for display SHALL already be stripped and normalized rather than preserving raw control-sequence text
