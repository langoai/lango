## MODIFIED Requirements

### Requirement: Approval status command
The system SHALL provide a `lango approval status [--json]` command that displays the current approval system status including approval mode, pending request count, and configured approval channels. The command SHALL use bootLoader because it reads approval provider state from the runtime.

#### Scenario: Approval command output uses the command writer
- **WHEN** `lango approval status` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
