## ADDED Requirements
### Requirement: Doctor CLI JSON output stays uniformly formatted
Pretty-printed JSON responses from the doctor CLI SHALL continue to flow through the shared CLI JSON writer path without changing payload shapes.

#### Scenario: Doctor CLI JSON payload shape remains unchanged
- **WHEN** `lango doctor --output json` renders its result
- **THEN** it SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
- **AND** the returned rendered string SHALL continue to omit the trailing newline
