## MODIFIED Requirements

### Requirement: Librarian inspection commands
The system SHALL provide `lango librarian status [--json]` and `lango librarian inquiries [--limit N] [--json]` commands for inspecting librarian configuration and pending inquiries.

#### Scenario: Librarian command output uses the command writer
- **WHEN** `lango librarian status` or `lango librarian inquiries` renders human-readable or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
