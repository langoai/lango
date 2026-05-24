## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Tool transcript rows keep detail previews width-safe
- **WHEN** a `tool` transcript row renders a preview or output/error detail line on a narrow terminal
- **THEN** each visible line SHALL clamp to the available transcript width instead of overflowing
- **AND** preview/output text SHALL be normalized to a single line before rendering
