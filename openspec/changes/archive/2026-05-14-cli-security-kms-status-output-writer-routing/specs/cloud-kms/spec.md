## MODIFIED Requirements

### Requirement: KMS status output routing
`lango security kms status` SHALL route human-readable and JSON output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: KMS status output uses the command writer
- **WHEN** `lango security kms status` renders text or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
