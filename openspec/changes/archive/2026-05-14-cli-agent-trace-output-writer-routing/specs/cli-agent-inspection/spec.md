## MODIFIED Requirements

### Requirement: Agent trace output routing
`lango agent trace list` and `lango agent trace show` SHALL route human-readable and JSON output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Agent trace output uses the command writer
- **WHEN** `lango agent trace` list or detail commands render text or JSON output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
