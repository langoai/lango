## MODIFIED Requirements

### Requirement: Observational memory inspection commands
The system SHALL provide `lango memory list --session <key>` and `lango memory status --session <key>` commands for inspecting observational memory entries and status.

#### Scenario: Memory command output uses the command writer
- **WHEN** `lango memory list` or `lango memory status` renders human-readable or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
