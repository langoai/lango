## ADDED Requirements
### Requirement: Memory CLI JSON output stays uniformly pretty-printed
Output-capable memory CLI subcommands SHALL continue to emit the same pretty-printed JSON shape through a shared package-local writer path.

#### Scenario: Memory JSON payload shape remains unchanged
- **WHEN** output-capable memory commands render `--json`
- **THEN** they SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
