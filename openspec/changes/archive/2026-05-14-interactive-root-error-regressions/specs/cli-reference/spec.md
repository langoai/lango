## MODIFIED Requirements

### Requirement: TUI startup notices remain seam-aware
Interactive top-level TUI entrypoints SHALL route their startup notice text through seam-aware stderr writers so wrapper and regression captures do not depend on process-global stderr interception.

#### Scenario: Cockpit rejects non-interactive startup cleanly
- **WHEN** `lango cockpit` is executed in a non-interactive environment
- **THEN** it SHALL return an actionable error requiring an interactive terminal

#### Scenario: Chat rejects non-interactive startup cleanly
- **WHEN** `lango chat` is executed in a non-interactive environment
- **THEN** it SHALL return an actionable error requiring an interactive terminal
