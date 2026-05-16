## ADDED Requirements
### Requirement: Smart-account output format values are validated before bootstrap work
`lango account` subcommands that expose `--output` SHALL accept only `table` or `json`. Unknown values SHALL fail before bootstrap or smart-account loading work begins.

#### Scenario: Deploy rejects an unknown output format before load
- **WHEN** the operator runs `lango account deploy --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke deploy loading work

#### Scenario: Session list rejects an unknown output format before load
- **WHEN** the operator runs `lango account session list --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke session-list loading work
