## MODIFIED Requirements

### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Approval-state help advertises immediate Ctrl+D quit
- **WHEN** the chat help bar is rendered in the `approving` state
- **THEN** it SHALL advertise `Ctrl+D` as the immediate quit path
