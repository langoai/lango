## ADDED Requirements
### Requirement: Workflow validate JSON output stays uniformly pretty-printed
Workflow validate JSON responses SHALL continue to emit the same pretty-printed JSON shape through a shared package-local writer path.

#### Scenario: Workflow validate JSON payload shape remains unchanged
- **WHEN** `lango workflow validate <file> --json` renders either success or validation-failure JSON
- **THEN** it SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
