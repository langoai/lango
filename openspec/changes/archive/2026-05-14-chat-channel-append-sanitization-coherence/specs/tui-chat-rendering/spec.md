## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Stored channel transcript payloads are sanitized
- **WHEN** the chat model appends a `channel` transcript row whose message text contains ANSI/OSC escape sequences
- **THEN** the stored transcript payload used for rendering SHALL already be stripped and normalized rather than preserving raw control-sequence text
