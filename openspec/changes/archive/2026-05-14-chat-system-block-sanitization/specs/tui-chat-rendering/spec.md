## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: System transcript blocks stay plain visible text
- **WHEN** a `system` transcript block renders content containing ANSI/OSC escape sequences
- **THEN** it SHALL strip those control sequences before displaying the block body
