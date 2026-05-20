## MODIFIED Requirements

### Requirement: Mission Control can project operator loops from real existing sources
Mission Control SHALL support an operator loop surface in addition to durable missions, proposals, and live decisions. In the first Slice 4 slice, loop rows SHALL be projected only from real existing sources rather than invented integrations or placeholder data.

#### Scenario: Scheduled-loop source text is replay-safe
- **WHEN** cron job names or last-run status text contain ANSI/OSC escape sequences or embedded newlines before loop projection
- **THEN** Mission Control SHALL strip those control sequences
- **AND** it SHALL normalize the scheduled-loop source text to a single line before projection
