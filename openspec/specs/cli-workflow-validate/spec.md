# CLI Workflow Validate

## Purpose
Provides a CLI command for validating YAML workflow definition files without executing them, checking syntax, required fields, dependency references, and DAG acyclicity.

## Requirements

### Requirement: Workflow validate command
The system SHALL provide a `lango workflow validate <file> [--output table|json]` command that parses and validates a YAML workflow definition file without executing it. The command SHALL check for valid YAML syntax, required fields (name, steps), step dependency references, and DAG acyclicity.

#### Scenario: Valid workflow file
- **WHEN** user runs `lango workflow validate workflow.yaml` with a well-formed workflow
- **THEN** system displays "Workflow 'name' is valid (N steps)"

#### Scenario: Invalid YAML syntax
- **WHEN** user runs `lango workflow validate broken.yaml` with malformed YAML
- **THEN** system returns error indicating YAML parse failure with line number

#### Scenario: Circular dependency
- **WHEN** user runs `lango workflow validate circular.yaml` with steps that form a cycle
- **THEN** system returns error "Workflow has circular dependencies"

#### Scenario: Missing step reference
- **WHEN** user runs `lango workflow validate missing-ref.yaml` where a step depends on a nonexistent step
- **THEN** system returns error indicating the unknown dependency reference

#### Scenario: Validate with JSON output
- **WHEN** user runs `lango workflow validate workflow.yaml --output json`
- **THEN** system outputs a JSON object with fields: valid, file, name, steps, and optional schedule

#### Scenario: Workflow validate rejects unknown output before parsing
- **WHEN** user runs `lango workflow validate workflow.yaml --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT parse the workflow file

#### Scenario: Workflow validate output uses the command writer
- **WHEN** `lango workflow validate` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

### Requirement: Workflow validate command registration
The `validate` subcommand SHALL be registered under the existing `lango workflow` command group.

#### Scenario: Workflow help lists validate
- **WHEN** user runs `lango workflow --help`
- **THEN** the help output includes validate in the available subcommands list
