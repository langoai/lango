## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Fullscreen approval dialog shows scroll help only when diff overflow exists
- **WHEN** the fullscreen approval dialog renders diff content that exceeds the visible diff area
- **THEN** its action bar SHALL advertise `↑/↓` scrolling

#### Scenario: Fullscreen approval dialog hides inert scroll help when diff fits
- **WHEN** the fullscreen approval dialog renders diff content that fits within the visible diff area
- **THEN** its action bar SHALL omit `↑/↓` scrolling
