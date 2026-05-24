## ADDED Requirements
### Requirement: RunLedger CLI output format stays explicit and validated
`lango run list`, `lango run status`, and `lango run journal` SHALL accept `--output table|json` and reject unknown values before bootstrap-dependent work.

#### Scenario: RunLedger CLI rejects unknown output before bootstrap
- **WHEN** the operator runs `lango run list`, `lango run status`, or `lango run journal <run-id>` with `--output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work
