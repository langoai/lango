## MODIFIED Requirements

### Requirement: Agent tools output routing
`lango agent tools` SHALL route human-readable and JSON output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Agent tools output uses the command writer
- **WHEN** `lango agent tools` renders text or JSON output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
