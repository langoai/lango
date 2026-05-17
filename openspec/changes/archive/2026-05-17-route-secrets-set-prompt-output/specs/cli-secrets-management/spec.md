## MODIFIED Requirements

### Requirement: Secrets command stream routing
`lango security secrets list`, `set`, and `delete` SHALL route human-readable and JSON output through the Cobra command writer. Interactive `set` and `delete` prompts SHALL write prompt text through the Cobra command output writer. Interactive delete confirmation SHALL read input through the Cobra command input reader.

#### Scenario: Secrets output uses the command writer
- **WHEN** `lango security secrets` commands render table, JSON, prompt, or success output
- **THEN** they SHALL write the full visible output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: Secrets set prompt uses command output
- **WHEN** `lango security secrets set <name>` prompts for an interactive secret value
- **THEN** it SHALL write the hidden-input prompt through the Cobra command output writer

#### Scenario: Secrets delete prompt uses command streams
- **WHEN** `lango security secrets delete <name>` prompts for confirmation
- **THEN** it SHALL write the prompt through the Cobra command output writer
- **AND** it SHALL read the operator response through the Cobra command input reader
