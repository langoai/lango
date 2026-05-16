## MODIFIED Requirements

### Requirement: Learning inspection commands
The system SHALL provide `lango learning status [--json]` and `lango learning history [--limit N] [--json]` commands for inspecting learning configuration and recent learning records.

#### Scenario: Learning command output uses the command writer
- **WHEN** `lango learning status` or `lango learning history` renders human-readable or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
