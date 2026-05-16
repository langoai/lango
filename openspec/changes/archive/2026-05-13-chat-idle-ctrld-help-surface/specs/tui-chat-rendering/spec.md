## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Idle and failed help advertise immediate Ctrl+D quit
- **WHEN** the chat help bar is rendered in the `idle` or `failed` states
- **THEN** it SHALL advertise `Ctrl+D` as the immediate quit path
