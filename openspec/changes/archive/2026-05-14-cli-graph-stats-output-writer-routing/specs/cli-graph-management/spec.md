## MODIFIED Requirements

### Requirement: Graph stats command
The system SHALL provide a `lango graph stats` command that displays total triple count and per-predicate breakdown sorted by count descending. The command SHALL support a `--json` flag.

#### Scenario: Graph stats output uses the command writer
- **WHEN** `lango graph stats` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
