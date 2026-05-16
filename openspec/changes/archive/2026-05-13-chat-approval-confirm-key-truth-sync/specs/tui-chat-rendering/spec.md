## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Approval confirm prompt reflects the pending action key
- **WHEN** a critical-risk approval surface is in confirm-pending state
- **THEN** the visible confirm prompt SHALL name the actual pending action key (`a` or `s`) rather than a hard-coded default
