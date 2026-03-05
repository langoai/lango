## ADDED Requirements

### Requirement: Payment x402 command
The system SHALL provide a `lango payment x402 [--json]` command that displays the X402 protocol configuration including enabled state, wallet address, payment endpoint, and accepted token types. The command SHALL use cfgLoader (config only).

#### Scenario: X402 enabled
- **WHEN** user runs `lango payment x402` with X402 enabled in configuration
- **THEN** system displays enabled state, wallet address, payment endpoint URL, and accepted tokens

#### Scenario: X402 disabled
- **WHEN** user runs `lango payment x402` with X402 disabled
- **THEN** system displays "X402 protocol is not enabled"

#### Scenario: X402 in JSON format
- **WHEN** user runs `lango payment x402 --json`
- **THEN** system outputs a JSON object with fields: enabled, walletAddress, endpoint, acceptedTokens

### Requirement: X402 command registration
The `x402` subcommand SHALL be registered under the existing `lango payment` command group.

#### Scenario: Payment help lists x402
- **WHEN** user runs `lango payment --help`
- **THEN** the help output includes x402 in the available subcommands list
