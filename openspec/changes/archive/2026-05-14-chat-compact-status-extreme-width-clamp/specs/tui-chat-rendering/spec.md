## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Compact status and approval rows clamp to extreme narrow widths
- **WHEN** a compact `status` or `approval` transcript row is rendered on an extremely narrow terminal
- **THEN** the final rendered row SHALL still clamp to the requested width
