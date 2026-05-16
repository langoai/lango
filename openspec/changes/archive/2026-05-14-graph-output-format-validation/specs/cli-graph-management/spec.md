## ADDED Requirements
### Requirement: Graph CLI output format stays explicit and validated
`lango graph status`, `query`, `stats`, `add`, and `import` SHALL accept `--output table|json` and reject unknown values before config loading or file parsing.

#### Scenario: Graph import rejects unknown output before file parsing
- **WHEN** user runs `lango graph import <file> --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT read or parse the import file
