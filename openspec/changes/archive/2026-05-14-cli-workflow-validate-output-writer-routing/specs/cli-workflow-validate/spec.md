## MODIFIED Requirements

### Requirement: Workflow validate command
The system SHALL provide a `lango workflow validate <file> [--json]` command that parses and validates a YAML workflow definition file without executing it. The command SHALL check for valid YAML syntax, required fields (name, steps), step dependency references, and DAG acyclicity. The command SHALL use cfgLoader for configuration access.

#### Scenario: Workflow validate output uses the command writer
- **WHEN** `lango workflow validate` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
