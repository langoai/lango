## ADDED Requirements
### Requirement: Status CLI pretty-JSON output stays uniformly formatted
Pretty-printed JSON responses from the status CLI SHALL continue to flow through the shared CLI JSON writer path without changing payload shapes.

#### Scenario: Status CLI pretty-JSON payload shape remains unchanged
- **WHEN** `lango status` renders JSON mode
- **THEN** it SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
