## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Recovery transcript rows keep action labels on the sanitized known-action path
- **WHEN** a `recovery` transcript row renders a known action value wrapped in ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences before mapping the action to its display label
