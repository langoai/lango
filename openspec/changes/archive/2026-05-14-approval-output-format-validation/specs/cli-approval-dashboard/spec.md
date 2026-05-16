## ADDED Requirements
### Requirement: Approval CLI output format stays explicit and validated
`lango approval status` SHALL accept `--output table|json` and reject unknown values before config loading.

#### Scenario: Approval status rejects unknown output before config load
- **WHEN** user runs `lango approval status --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader
