## MODIFIED Requirements

### Requirement: Mission Control defines loading, empty, degraded, and responsive states
Mission Control SHALL distinguish first-load, empty-data, degraded-reader, and narrow-terminal states instead of rendering the same fallback for every case.

#### Scenario: Rendered Mission Control lane text stays plain and single-line
- **WHEN** mission titles/details, collaboration labels, proposal source labels, decision text, activity summaries, loop titles/details, or lane overflow summaries contain ANSI/OSC escape sequences or embedded newlines
- **THEN** Mission Control SHALL strip those control sequences
- **AND** it SHALL normalize the displayed lane text to a single line before rendering it
