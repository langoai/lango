## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Fullscreen approval dialog surfaces split-toggle help when diff exists
- **WHEN** the fullscreen approval dialog renders diff content
- **THEN** its action bar SHALL advertise the `t` key for toggling split mode
