## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Thinking transcript rows stay plain and single-line
- **WHEN** a `thinking` transcript row renders preview or fallback text containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before truncating it into the row
