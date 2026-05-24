## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Channel transcript rows sanitize remote sender and message text
- **WHEN** a `channel` transcript row renders remote sender or message text containing ANSI/OSC escape sequences
- **THEN** it SHALL strip those control sequences before display
- **AND** the rendered row SHALL remain a plain visible text surface
