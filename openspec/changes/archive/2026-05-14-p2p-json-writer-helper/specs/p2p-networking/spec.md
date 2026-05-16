## ADDED Requirements
### Requirement: P2P CLI JSON output stays uniformly pretty-printed
Output-capable P2P CLI subcommands SHALL continue to emit the same pretty-printed JSON shape through a shared package-local writer path.

#### Scenario: P2P JSON payload shape remains unchanged
- **WHEN** output-capable P2P commands render `--json`
- **THEN** they SHALL still emit valid pretty-printed JSON payloads matching the existing field shapes
