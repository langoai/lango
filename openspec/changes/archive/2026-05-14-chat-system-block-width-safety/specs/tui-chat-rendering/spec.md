## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: System transcript blocks stay width-safe
- **WHEN** a `system` transcript item is rendered on a narrow terminal
- **THEN** the rendered block SHALL clamp to the available width instead of overflowing the transcript layout
