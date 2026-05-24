## ADDED Requirements
### Requirement: A2A CLI JSON output stays uniformly pretty-printed
Output-capable A2A CLI subcommands SHALL continue to emit the same pretty-printed JSON shape through a shared package-local writer path.

#### Scenario: A2A JSON payload shape remains unchanged
- **WHEN** output-capable A2A commands render `--json`
- **THEN** they SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
