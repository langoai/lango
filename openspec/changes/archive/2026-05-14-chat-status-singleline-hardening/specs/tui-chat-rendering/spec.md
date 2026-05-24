## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Compact status and approval rows stay single-line safe
- **WHEN** a compact `status` or `approval` transcript row is rendered from content containing line breaks
- **THEN** the rendered row SHALL collapse those line breaks instead of emitting multiline compact rows
