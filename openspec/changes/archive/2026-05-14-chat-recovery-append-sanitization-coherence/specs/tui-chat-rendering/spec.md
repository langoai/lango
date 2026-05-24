## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Stored recovery transcript cause metadata is sanitized
- **WHEN** the chat model appends a `recovery` transcript row whose `causeClass` contains ANSI/OSC escape sequences or embedded newlines
- **THEN** the stored recovery metadata SHALL already be stripped and normalized rather than preserving raw control-sequence text
