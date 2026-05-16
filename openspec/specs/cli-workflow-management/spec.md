## Purpose

Define the CLI commands for managing workflow execution (run, list, status, cancel, history).
## Requirements
### Requirement: Workflow run command
The CLI SHALL provide `lango workflow run <file.yaml>` that parses and executes a workflow YAML file.

#### Scenario: Run a workflow
- **WHEN** user runs `lango workflow run code-review.flow.yaml`
- **THEN** the CLI SHALL parse the YAML, validate the DAG, and execute the workflow

#### Scenario: Run with schedule registration
- **WHEN** user runs `lango workflow run report.flow.yaml --schedule "0 9 * * MON"`
- **THEN** the CLI SHALL validate and display the workflow plus the effective schedule
- **AND** SHALL report that CLI schedule registration is not implemented yet
- **AND** SHALL direct the operator toward `lango cron add` or runtime automation tools instead of a nonexistent workflow registration endpoint

#### Scenario: Run without runtime reports server unavailable guidance
- **WHEN** user runs `lango workflow run code-review.flow.yaml`
- **AND** the live runtime bootstrap is unavailable
- **THEN** the CLI SHALL still display the validated workflow summary
- **AND** SHALL report that direct execution is not available from the current runtime state

#### Scenario: Run with workflow engine disabled reports guidance
- **WHEN** user runs `lango workflow run code-review.flow.yaml`
- **AND** bootstrap succeeds but the workflow engine is disabled in config
- **THEN** the CLI SHALL still display the validated workflow summary
- **AND** SHALL report that the workflow engine is not enabled

#### Scenario: Run success writes completion details to command output
- **WHEN** user runs `lango workflow run code-review.flow.yaml`
- **AND** direct execution succeeds
- **THEN** the CLI SHALL write the execution banner, final status, and per-step output to the Cobra command output stream

#### Scenario: Invalid YAML file
- **WHEN** user runs `lango workflow run invalid.yaml` with malformed content
- **THEN** the CLI SHALL display a parse error

### Requirement: Workflow run output routing
`lango workflow run` SHALL write its validation, schedule guidance, and direct execution status output through the Cobra command output stream so wrappers and test harnesses can capture non-error output without intercepting process-global stdout.

#### Scenario: Workflow run schedule guidance writes to command output
- **WHEN** user runs `lango workflow run report.flow.yaml --schedule "0 9 * * MON"`
- **THEN** the validated workflow summary and schedule guidance write to the Cobra command output stream

### Requirement: Workflow list command
The CLI SHALL provide `lango workflow list` that displays workflow runs with columns: Run ID, Workflow, Status, Steps, Started, Completed.

#### Scenario: List workflow runs
- **WHEN** user runs `lango workflow list`
- **THEN** the CLI SHALL display all workflow runs in tabular format

### Requirement: Workflow status command
The CLI SHALL provide `lango workflow status <run-id>` that displays detailed run information including all step statuses.

#### Scenario: View workflow status
- **WHEN** user runs `lango workflow status <uuid>`
- **THEN** the CLI SHALL display the run overview and a table of step statuses

### Requirement: Workflow cancel command
The CLI SHALL provide `lango workflow cancel <run-id>` that cancels a running workflow.

#### Scenario: Cancel a running workflow
- **WHEN** user runs `lango workflow cancel <uuid>`
- **THEN** the CLI SHALL cancel the workflow and display confirmation

### Requirement: Workflow history command
The CLI SHALL provide `lango workflow history` that displays completed workflow runs.

#### Scenario: View workflow history
- **WHEN** user runs `lango workflow history`
- **THEN** the CLI SHALL display recent workflow runs ordered by start time

### Requirement: Workflow management output routing
`lango workflow list`, `status`, `history`, and `cancel` SHALL write non-error output through the Cobra command output stream so wrappers and test harnesses can capture table, detail, empty-state, and confirmation output without intercepting process-global stdout.

#### Scenario: Workflow list writes to command output
- **WHEN** user runs `lango workflow list`
- **THEN** the list table or empty-state message writes to the Cobra command output stream

#### Scenario: Workflow status writes to command output
- **WHEN** user runs `lango workflow status <uuid>`
- **THEN** the status detail output writes to the Cobra command output stream

#### Scenario: Workflow history writes to command output
- **WHEN** user runs `lango workflow history`
- **THEN** the history table or empty-state message writes to the Cobra command output stream

#### Scenario: Workflow cancel writes to command output
- **WHEN** user runs `lango workflow cancel <uuid>`
- **THEN** the cancellation confirmation writes to the Cobra command output stream

### Requirement: Validate subcommand addition
The existing `lango workflow` command group SHALL gain a new `validate` subcommand that validates YAML workflow definition files without executing them. This extends the workflow CLI surface without modifying existing workflow subcommands.

#### Scenario: Workflow help includes validate
- **WHEN** user runs `lango workflow --help`
- **THEN** the help output lists validate alongside existing workflow subcommands (list, run, status)

### Requirement: Validate subcommand uses cfgLoader
The `workflow validate` subcommand SHALL use cfgLoader to access workflow engine configuration. It SHALL NOT require bootLoader or full workflow engine initialization.

#### Scenario: Config-only access
- **WHEN** user runs `lango workflow validate workflow.yaml`
- **THEN** the command loads configuration via cfgLoader and parses the YAML file against the workflow schema

### Requirement: Validate does not execute
The validate subcommand SHALL only parse and check the workflow definition. It SHALL NOT execute any steps, connect to external services, or modify any state.

#### Scenario: No side effects
- **WHEN** user runs `lango workflow validate workflow.yaml`
- **THEN** no workflow steps are executed and no database writes occur

### Requirement: Existing workflow commands unaffected
The addition of the validate subcommand SHALL NOT change the behavior or registration of any existing workflow subcommands.

#### Scenario: Existing commands still work
- **WHEN** user runs existing `lango workflow list` command
- **THEN** the command behaves identically to before the validate addition

### Requirement: Workflow CLI uses workflow state store capability
Workflow CLI commands MUST obtain workflow state persistence through a storage facade capability instead of constructing a state store from a generic Ent client in the CLI layer.

#### Scenario: Workflow engine initialization uses facade state store
- **WHEN** the workflow CLI initializes a workflow engine
- **THEN** it resolves the workflow state store from the storage facade capability

### Requirement: Workflow history/status support broker-backed runtime reads
Workflow CLI read surfaces MUST remain functional when runtime state is served by broker-backed storage.

#### Scenario: Workflow read path under broker-owned runtime
- **WHEN** broker-backed runtime storage is active
- **THEN** workflow list/history/status read state through broker-backed storage capabilities
