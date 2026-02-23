## ADDED Requirements

### Requirement: CLI pricing command
The system SHALL provide a `lango p2p pricing` CLI command that displays P2P tool pricing configuration.

#### Scenario: Show all pricing
- **WHEN** user runs `lango p2p pricing`
- **THEN** system displays enabled status, default per-query price, and tool-specific price overrides in table format

#### Scenario: Show pricing for specific tool
- **WHEN** user runs `lango p2p pricing --tool "knowledge_search"`
- **THEN** system displays the price for that specific tool (or default per-query price if no override)

#### Scenario: Show pricing as JSON
- **WHEN** user runs `lango p2p pricing --json`
- **THEN** system outputs full pricing config as JSON to stdout

### Requirement: CLI pricing registered as subcommand
The `pricing` command SHALL be registered as a subcommand of `lango p2p` in `internal/cli/p2p/p2p.go`.

#### Scenario: Help shows pricing command
- **WHEN** user runs `lango p2p --help`
- **THEN** output lists `pricing` as an available subcommand
