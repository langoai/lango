## MODIFIED Requirements

### Requirement: Mission Control defines loading, empty, degraded, and responsive states
Mission Control SHALL distinguish first-load, empty-data, degraded-reader, and narrow-terminal states instead of rendering the same fallback for every case.

#### Scenario: Buffered activity summaries are replay-safe
- **WHEN** assistant or runtime activity summaries contain ANSI/OSC escape sequences or embedded newlines before buffering
- **THEN** the Mission Control activity buffer SHALL strip those control sequences
- **AND** it SHALL normalize the stored summary text to a single line before replay
