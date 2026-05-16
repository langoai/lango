## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Idle and failed help describe double-press quit
- **WHEN** the chat help bar is rendered in the `idle` or `failed` states
- **THEN** the `Ctrl+C` binding SHALL describe quitting via the double-press path rather than immediate single-press exit
