## ADDED Requirements
### Requirement: Memory CLI output format stays explicit and validated
`lango memory list`, `lango memory status`, `lango memory agents`, and `lango memory agent` SHALL accept `--output table|json` and reject unknown values before config loading.

#### Scenario: Memory list rejects unknown output before config load
- **WHEN** user runs `lango memory list --session my-session --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

#### Scenario: Memory status rejects unknown output before config load
- **WHEN** user runs `lango memory status --session my-session --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

#### Scenario: Memory agents reject unknown output before config load
- **WHEN** user runs `lango memory agents --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

#### Scenario: Memory agent rejects unknown output before config load
- **WHEN** user runs `lango memory agent <name> --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader
