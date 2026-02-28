## ADDED Requirements

### Requirement: bg command is registered in main.go
The `lango bg` command SHALL be registered in `cmd/lango/main.go` with GroupID "infra", using a stub manager provider that returns an error when invoked outside a running server.

#### Scenario: bg command appears in help
- **WHEN** user runs `lango --help`
- **THEN** the `bg` command SHALL appear under the "Infrastructure" group

#### Scenario: bg subcommand returns server-required error
- **WHEN** user runs `lango bg list` without a running server
- **THEN** the command SHALL return an error containing "bg commands require a running server"
