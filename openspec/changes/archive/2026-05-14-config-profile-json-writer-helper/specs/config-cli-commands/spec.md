## ADDED Requirements
### Requirement: Config profile export JSON stays uniformly pretty-printed
`lango config export <name>` SHALL continue to emit the same pretty-printed JSON shape through the shared CLI JSON writer path.

#### Scenario: Config export JSON payload shape remains unchanged
- **WHEN** `lango config export <name>` renders the exported profile JSON
- **THEN** it SHALL still emit valid pretty-printed JSON matching the existing field shapes
