## ADDED Requirements
### Requirement: Extension CLI pretty-JSON output stays uniformly formatted
Pretty-printed JSON responses from extension CLI inspect and list surfaces SHALL continue to flow through a shared package-local writer path without changing payload shapes.

#### Scenario: Extension CLI pretty-JSON payload shapes remain unchanged
- **WHEN** extension CLI inspect or list commands render JSON mode
- **THEN** they SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
