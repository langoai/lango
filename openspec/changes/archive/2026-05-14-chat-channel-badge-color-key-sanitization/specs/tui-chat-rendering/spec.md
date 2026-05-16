## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Channel badge color follows the sanitized channel name
- **WHEN** a `channel` transcript row receives a known channel name wrapped in ANSI/OSC escape sequences
- **THEN** the badge text SHALL use the sanitized visible name
- **AND** the badge color selection SHALL use that same sanitized channel name rather than the raw control-sequence input
