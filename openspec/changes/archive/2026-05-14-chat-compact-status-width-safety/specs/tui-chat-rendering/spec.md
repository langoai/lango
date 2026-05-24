## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Compact status and approval rows stay width-safe
- **WHEN** a compact `status` or `approval` transcript row is rendered on a narrow terminal
- **THEN** the final rendered row SHALL truncate to the available width instead of overflowing the transcript layout
