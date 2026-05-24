## ADDED Requirements
### Requirement: Workflow validate output format stays explicit and validated
`lango workflow validate` SHALL accept `--output table|json` and reject unknown values before parsing the workflow file.

#### Scenario: Workflow validate rejects unknown output before parsing
- **WHEN** user runs `lango workflow validate workflow.yaml --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT parse the workflow file
