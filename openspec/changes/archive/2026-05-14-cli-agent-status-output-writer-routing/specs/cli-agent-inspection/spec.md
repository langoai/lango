## MODIFIED Requirements

### Requirement: Agent status output routing
`lango agent status` SHALL route human-readable and JSON output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Agent status output uses the command writer
- **WHEN** `lango agent status` renders table or JSON output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
