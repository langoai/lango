## MODIFIED Requirements

### Requirement: Mission Control defines loading, empty, degraded, and responsive states
Mission Control SHALL distinguish first-load, empty-data, degraded-reader, and narrow-terminal states instead of rendering the same fallback for every case.

#### Scenario: Empty-state last result stays plain and single-line
- **WHEN** the latest assistant activity summary shown as `Last result:` contains ANSI/OSC escape sequences or embedded newlines
- **THEN** Mission Control SHALL strip those control sequences
- **AND** it SHALL normalize the displayed empty-state result text to a single line before rendering it
