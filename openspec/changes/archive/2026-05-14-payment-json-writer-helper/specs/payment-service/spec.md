## ADDED Requirements
### Requirement: Payment CLI JSON output stays uniformly pretty-printed
Output-capable payment CLI subcommands SHALL continue to emit the same pretty-printed JSON shape through a shared package-local writer path.

#### Scenario: Payment JSON payload shape remains unchanged
- **WHEN** output-capable payment commands render `--json`
- **THEN** they SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
