## MODIFIED Requirements

### Requirement: Graph status command
The system SHALL provide a `lango graph status` command that displays the graph store configuration and triple count. The command SHALL support a `--json` flag for machine-readable output.

#### Scenario: Graph status output uses the command writer
- **WHEN** `lango graph status` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
