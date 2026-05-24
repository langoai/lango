## ADDED Requirements
### Requirement: Agent CLI pretty-JSON output stays uniformly formatted
Pretty-printed JSON responses from agent CLI diagnostics subcommands SHALL continue to flow through a shared package-local writer path without changing payload shapes.

#### Scenario: Agent CLI pretty-JSON payload shape remains unchanged
- **WHEN** output-capable agent diagnostics commands render their pretty-printed JSON mode
- **THEN** they SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
