## ADDED Requirements
### Requirement: Security CLI JSON output stays uniformly pretty-printed
Output-capable security CLI subcommands SHALL continue to emit the same pretty-printed JSON shape through a shared package-local writer path.

#### Scenario: Security JSON payload shape remains unchanged
- **WHEN** output-capable security commands render `--json`
- **THEN** they SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
