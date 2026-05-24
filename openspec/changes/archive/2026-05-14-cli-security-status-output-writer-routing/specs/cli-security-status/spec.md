## MODIFIED Requirements

### Requirement: Security status output routing
`lango security status` SHALL route human-readable and JSON output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Security status output uses the command writer
- **WHEN** `lango security status` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
