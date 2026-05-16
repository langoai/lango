# CLI Approval Dashboard

## Purpose
Provides CLI commands for viewing the approval system status, including approval mode, pending request count, and configured approval channels.

## Requirements

### Requirement: Approval status command
The system SHALL provide a `lango approval status [--output table|json]` command that displays the current approval system status including approval mode, pending request count, and configured approval channels. The command SHALL use cfgLoader.

#### Scenario: Approval enabled
- **WHEN** user runs `lango approval status` with approval system enabled
- **THEN** system displays approval mode (auto/manual/channel), pending request count, and configured approval channels

#### Scenario: Approval disabled
- **WHEN** user runs `lango approval status` with approval system disabled
- **THEN** system displays "Approval system is disabled"

#### Scenario: Approval status in JSON format
- **WHEN** user runs `lango approval status --output json`
- **THEN** system outputs a JSON object with approval-interceptor status fields

#### Scenario: Approval status rejects unknown output before config load
- **WHEN** user runs `lango approval status --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

#### Scenario: Approval command output uses the command writer
- **WHEN** `lango approval status` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

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
