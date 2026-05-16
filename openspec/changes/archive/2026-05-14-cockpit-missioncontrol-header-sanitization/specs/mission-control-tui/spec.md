## MODIFIED Requirements

### Requirement: Mission Control presents timeline and header as first-class Wave 1 outputs
Mission Control SHALL add a real direct mission-start write path in Wave 2 while preserving timeline and header behavior from Wave 1.

#### Scenario: Rendered Mission Control header text stays plain and single-line
- **WHEN** active-agent, model/provider, context, metrics, or degraded-note header summaries contain ANSI/OSC escape sequences or embedded newlines
- **THEN** Mission Control SHALL strip those control sequences
- **AND** it SHALL normalize the displayed header text to a single line before rendering it
