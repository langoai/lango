## ADDED Requirements
### Requirement: Librarian CLI output format stays explicit and validated
`lango librarian status` and `lango librarian inquiries` SHALL accept `--output table|json` and reject unknown values before config or bootstrap loading.

#### Scenario: Librarian status rejects unknown output before config load
- **WHEN** user runs `lango librarian status --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

#### Scenario: Librarian inquiries rejects unknown output before bootstrap
- **WHEN** user runs `lango librarian inquiries --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the bootstrap loader
