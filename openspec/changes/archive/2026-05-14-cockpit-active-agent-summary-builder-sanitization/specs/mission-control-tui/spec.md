## MODIFIED Requirements

### Requirement: Mission Control presents timeline and header as first-class Wave 1 outputs
Mission Control SHALL add a real direct mission-start write path in Wave 2 while preserving timeline and header behavior from Wave 1.

#### Scenario: Active-agent summary aggregation stays replay-safe
- **WHEN** mission owner-agent labels contain ANSI/OSC escape sequences or embedded newlines before header aggregation
- **THEN** the Mission Control active-agent summary builder SHALL strip those control sequences
- **AND** it SHALL normalize the aggregated header summary text to a single line before replay
