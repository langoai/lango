## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Approval confirm prompt keeps deny path visible
- **WHEN** a critical-risk approval surface is in confirm-pending state
- **THEN** the visible confirm prompt SHALL mention that `d` or `Esc` still denies the request
