## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Stored tool transcript names are sanitized
- **WHEN** the chat model appends a tool transcript row whose tool name contains ANSI/OSC escape sequences or embedded newlines
- **THEN** the stored transcript entry name SHALL already be stripped and normalized rather than preserving raw control-sequence text
