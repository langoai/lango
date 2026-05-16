## MODIFIED Requirements

### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Approval-state turn strip uses unified deny-key wording
- **WHEN** the turn status strip renders in the `approving` state
- **THEN** it SHALL surface the deny path using the shared `d/Esc` wording
