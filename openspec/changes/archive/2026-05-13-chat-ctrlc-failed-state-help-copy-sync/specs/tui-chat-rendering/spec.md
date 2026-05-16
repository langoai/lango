## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Help text describes double-press quit for idle and failed states
- **WHEN** a user reads the chat help copy for `Ctrl+C`
- **THEN** it SHALL describe the double-press quit path for both `idle` and `failed` states
