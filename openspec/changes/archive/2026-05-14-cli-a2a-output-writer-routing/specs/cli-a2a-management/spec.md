## MODIFIED Requirements

### Requirement: A2A card command
The system SHALL provide a `lango a2a card [--json]` command that displays the local agent's A2A agent card including name, description, capabilities, and endpoint URL. The command SHALL use cfgLoader to read the A2A configuration.

#### Scenario: A2A CLI output uses the command writer
- **WHEN** `lango a2a card` or `lango a2a check` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
