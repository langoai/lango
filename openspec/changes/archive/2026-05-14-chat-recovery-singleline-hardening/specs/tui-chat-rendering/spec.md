## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Recovery transcript rows stay compact and single-line
- **WHEN** a `recovery` transcript row renders multiline or whitespace-heavy recovery metadata
- **THEN** the row SHALL normalize that metadata to a single line
- **AND** the rendered row SHALL remain a compact one-line event
