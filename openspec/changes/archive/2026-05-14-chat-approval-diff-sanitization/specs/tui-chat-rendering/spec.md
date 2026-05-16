## MODIFIED Requirements

### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Fullscreen approval dialog sanitizes diff preview text
- **WHEN** the fullscreen approval dialog renders diff content containing ANSI/OSC escape sequences
- **THEN** it SHALL strip those control sequences before styling or displaying the diff lines
