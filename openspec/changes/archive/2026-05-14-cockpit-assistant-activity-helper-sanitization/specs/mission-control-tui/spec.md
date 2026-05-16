## MODIFIED Requirements

### Requirement: Mission Control defines loading, empty, degraded, and responsive states
Mission Control SHALL distinguish first-load, empty-data, degraded-reader, and narrow-terminal states instead of rendering the same fallback for every case.

#### Scenario: Exported assistant activity helper is replay-safe
- **WHEN** assistant summary, response text, or user message fields contain ANSI/OSC escape sequences or embedded newlines before helper construction
- **THEN** `NewAssistantSummaryActivity()` SHALL strip those control sequences
- **AND** it SHALL normalize the returned summary text to a single line before replay
