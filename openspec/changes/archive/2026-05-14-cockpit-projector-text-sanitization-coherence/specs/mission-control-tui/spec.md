## MODIFIED Requirements

### Requirement: Mission Control defines loading, empty, degraded, and responsive states
Mission Control SHALL distinguish first-load, empty-data, degraded-reader, and narrow-terminal states instead of rendering the same fallback for every case.

#### Scenario: Projected Mission Control snapshot text is replay-safe
- **WHEN** proposal, decision, collaboration, loop, or durable-mission text contains ANSI/OSC escape sequences or embedded newlines before projection
- **THEN** the Mission Control projector SHALL strip those control sequences
- **AND** it SHALL normalize the stored snapshot text to a single line before replay
