## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Pending indicator stays width-safe
- **WHEN** the submit-to-first-event pending indicator renders on a narrow terminal
- **THEN** it SHALL clamp to the available transcript width instead of overflowing
