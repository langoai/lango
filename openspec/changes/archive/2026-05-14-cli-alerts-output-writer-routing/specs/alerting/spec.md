## MODIFIED Requirements

### Requirement: CLI alerts list command
The system SHALL provide a `lango alerts list` CLI command that queries the gateway `/alerts` endpoint and displays recent alerts. The command SHALL support `--days` flag (default: 7) and `--output table|json` format.

#### Scenario: Alerts command output uses the command writer
- **WHEN** `lango alerts list` or `lango alerts summary` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
