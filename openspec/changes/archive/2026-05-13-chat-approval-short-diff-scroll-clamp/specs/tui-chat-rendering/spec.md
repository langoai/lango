## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Fullscreen approval diff clamps scroll offset when the diff fits
- **WHEN** the fullscreen approval dialog renders a diff that fits within the visible diff area
- **THEN** its scroll offset SHALL clamp to zero
- **AND** the full short diff SHALL remain visible
