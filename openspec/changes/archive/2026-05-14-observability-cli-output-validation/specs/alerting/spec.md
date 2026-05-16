## ADDED Requirements
### Requirement: Alerts output format values are validated before gateway calls
`lango alerts` subcommands SHALL accept only `table` or `json` for `--output`. Unknown values SHALL fail before the command contacts the gateway.

#### Scenario: Alerts list rejects an unknown output format before fetch
- **WHEN** the operator runs `lango alerts list --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT contact the gateway
