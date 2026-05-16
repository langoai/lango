## ADDED Requirements

### Requirement: P2P pricing command

The system SHALL provide `lango p2p pricing [--tool <name>] [--json]` that displays provider-side P2P quote configuration including the default per-query price and tool-specific quote overrides.

#### Scenario: Pricing overview in text format
- **WHEN** user runs `lango p2p pricing`
- **THEN** the command prints whether pricing is enabled, the default per-query price, and any tool-specific prices

#### Scenario: Single tool pricing in text format
- **WHEN** user runs `lango p2p pricing --tool knowledge_search`
- **THEN** the command prints the selected tool and its public quote

#### Scenario: Pricing in JSON format
- **WHEN** user runs `lango p2p pricing --json`
- **THEN** the JSON output SHALL include `enabled`, `perQuery`, `toolPrices`, and `currency`

### Requirement: P2P pricing output routing

`lango p2p pricing` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON output without intercepting process-global stdout.

#### Scenario: Pricing text output writes to command output
- **WHEN** user runs `lango p2p pricing`
- **THEN** the command writes the pricing overview to the Cobra command output stream

#### Scenario: Pricing tool output writes to command output
- **WHEN** user runs `lango p2p pricing --tool knowledge_search`
- **THEN** the command writes the tool-specific price view to the Cobra command output stream

#### Scenario: Pricing JSON output writes to command output
- **WHEN** user runs `lango p2p pricing --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream
