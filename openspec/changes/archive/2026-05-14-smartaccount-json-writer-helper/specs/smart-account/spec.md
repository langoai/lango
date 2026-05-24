## ADDED Requirements
### Requirement: Smart-account JSON output stays uniformly pretty-printed
Output-capable smart-account CLI subcommands SHALL continue to emit the same pretty-printed JSON shape through a shared package-local writer path.

#### Scenario: Smart-account JSON payload shape remains unchanged
- **WHEN** output-capable smart-account commands render `--output json`
- **THEN** they SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
