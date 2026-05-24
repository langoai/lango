## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: User transcript blocks stay plain visible text
- **WHEN** a user transcript block renders submitted content containing ANSI/OSC escape sequences
- **THEN** it SHALL strip those control sequences before displaying the prompt text
