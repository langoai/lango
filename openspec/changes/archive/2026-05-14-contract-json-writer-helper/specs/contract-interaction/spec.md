## ADDED Requirements
### Requirement: Contract CLI JSON output stays uniformly pretty-printed
Output-capable contract CLI subcommands SHALL continue to emit the same pretty-printed JSON shape through a shared package-local writer path.

#### Scenario: Contract JSON payload shape remains unchanged
- **WHEN** output-capable contract commands render `--output`
- **THEN** they SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
