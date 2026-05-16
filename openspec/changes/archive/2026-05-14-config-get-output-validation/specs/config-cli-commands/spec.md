## ADDED Requirements
### Requirement: Config get output format values are validated before loading config
`lango config get` SHALL accept only `plain` or `json` for `--output`. Unknown values SHALL fail before the command loads the active config.

#### Scenario: Config get rejects an unknown output format before load
- **WHEN** `lango config get <path> --output yaml` is run
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT load the active config
