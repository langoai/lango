## MODIFIED Requirements

### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Approval-state strip hint reflects deny and quit affordances
- **WHEN** the turn state strip is rendered in the `approving` state
- **THEN** its hint SHALL mention the `a` and `s` approval paths
- **AND** SHALL mention `d/Esc` for denial
- **AND** SHALL mention `Ctrl+D` as the immediate quit path
