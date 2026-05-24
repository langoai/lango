## MODIFIED Requirements

### Requirement: Mission Control presents timeline and header as first-class Slice 1 outputs
Mission Control SHALL add a real direct mission-start write path in Slice 2 while preserving timeline and header behavior from Slice 1.

#### Scenario: Active-agent summary aggregation stays replay-safe
- **WHEN** mission owner-agent labels contain ANSI/OSC escape sequences or embedded newlines before header aggregation
- **THEN** the Mission Control active-agent summary builder SHALL strip those control sequences
- **AND** it SHALL normalize the aggregated header summary text to a single line before replay
