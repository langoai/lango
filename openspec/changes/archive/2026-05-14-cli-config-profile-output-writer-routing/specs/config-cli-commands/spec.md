## MODIFIED Requirements

### Requirement: Config create command
The system SHALL provide a `lango config create <name>` command that creates a new profile with default configuration.

#### Scenario: Config profile-management output uses the command writer
- **WHEN** `lango config list`, `create`, `use`, `delete`, or `import` renders output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** `delete` confirmation SHALL read from `cmd.InOrStdin()`
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` / `cmd.InOrStdin()` SHALL capture the interaction
