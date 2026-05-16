## MODIFIED Requirements

### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Fullscreen approval badge color follows the sanitized risk level
- **WHEN** the fullscreen approval dialog receives a known risk level wrapped in ANSI/OSC escape sequences
- **THEN** the badge text SHALL use the sanitized visible level
- **AND** the badge color selection SHALL use that same sanitized risk level rather than the raw control-sequence input
