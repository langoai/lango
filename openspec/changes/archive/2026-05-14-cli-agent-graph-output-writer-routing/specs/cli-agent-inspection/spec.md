## MODIFIED Requirements

### Requirement: Agent graph output routing
`lango agent graph` SHALL route human-readable and JSON output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Agent graph output uses the command writer
- **WHEN** `lango agent graph` renders text or JSON output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
