## ADDED Requirements

### Requirement: Approval status command
The system SHALL provide a `lango approval status [--json]` command that displays the current approval system status including approval mode, pending request count, and configured approval channels. The command SHALL use bootLoader because it reads approval provider state from the runtime.

#### Scenario: Approval enabled
- **WHEN** user runs `lango approval status` with approval system enabled
- **THEN** system displays approval mode (auto/manual/channel), pending request count, and configured approval channels

#### Scenario: Approval disabled
- **WHEN** user runs `lango approval status` with approval system disabled
- **THEN** system displays "Approval system is disabled"

#### Scenario: Approval status in JSON format
- **WHEN** user runs `lango approval status --json`
- **THEN** system outputs a JSON object with fields: enabled, mode, pendingCount, channels

### Requirement: Approval command entry point
The system SHALL provide a `lango approval` command group. When invoked without a subcommand, it SHALL display help text listing the status subcommand.

#### Scenario: Help text
- **WHEN** user runs `lango approval`
- **THEN** system displays help listing the status subcommand

### Requirement: Approval command registration
The `approval` command group SHALL be registered in `cmd/lango/main.go` as a top-level command group.

#### Scenario: Root help includes approval
- **WHEN** user runs `lango --help`
- **THEN** the help output includes the approval command in the list of available commands
