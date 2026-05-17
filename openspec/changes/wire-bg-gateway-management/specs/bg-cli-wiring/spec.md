## MODIFIED Requirements

### Requirement: bg command is registered in main.go
The `lango bg` command SHALL be registered in `cmd/lango/main.go` with GroupID "auto", using a gateway-backed client for root CLI execution while preserving in-process manager adapters for embedded callers.

#### Scenario: bg command appears in help
- **WHEN** user runs `lango --help`
- **THEN** the `bg` command SHALL appear under the "Automation" group

#### Scenario: Root CLI background command uses gateway management
- **WHEN** the user runs a root `lango bg` subcommand
- **THEN** the command SHALL use the configured gateway background management API instead of returning the standalone in-memory boundary stub
- **AND** the command SHALL still explain gateway connection failures as failures to reach the running Lango gateway

#### Scenario: In-process background command behavior remains available
- **WHEN** `internal/cli/bg.NewBgCmd` is constructed with a real in-process background client
- **THEN** `list`, `status`, `cancel`, and `result` SHALL continue to operate on that in-process manager
