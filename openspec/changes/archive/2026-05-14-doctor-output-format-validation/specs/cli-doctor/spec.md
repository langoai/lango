## ADDED Requirements
### Requirement: Doctor output format stays explicit and validated
`lango doctor` SHALL accept `--output table|json` and reject unknown values before bootstrap.

#### Scenario: Doctor rejects unknown output before bootstrap
- **WHEN** `lango doctor --output yaml` is run
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap
