## MODIFIED Requirements

### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Fullscreen approval dialog keeps risk text plain and single-line
- **WHEN** the fullscreen approval dialog renders risk-label or rule-explanation text containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed risk text to plain single-line text before rendering it
