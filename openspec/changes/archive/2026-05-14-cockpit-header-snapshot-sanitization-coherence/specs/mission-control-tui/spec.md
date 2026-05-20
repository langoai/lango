## MODIFIED Requirements

### Requirement: Mission Control presents timeline and header as first-class Slice 1 outputs
Mission Control SHALL add a real direct mission-start write path in Slice 2 while preserving timeline and header behavior from Slice 1.

#### Scenario: Projected Mission Control header snapshot text is replay-safe
- **WHEN** active-agent, provider/model, metrics, or degraded-note header summaries contain ANSI/OSC escape sequences or embedded newlines before projection
- **THEN** the Mission Control projector SHALL strip those control sequences
- **AND** it SHALL normalize the stored header snapshot text to a single line before replay
