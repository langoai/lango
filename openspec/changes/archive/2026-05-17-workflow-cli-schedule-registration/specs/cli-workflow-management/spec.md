## MODIFIED Requirements

### Requirement: Workflow run command
The CLI SHALL provide `lango workflow run <file.yaml>` that parses and executes a workflow YAML file.

#### Scenario: Run with schedule registration
- **WHEN** user runs `lango workflow run report.flow.yaml --schedule "0 9 * * MON"`
- **THEN** the CLI SHALL validate and display the workflow plus the effective schedule
- **AND** SHALL register an enabled cron job for the workflow through the runtime cron storage facade
- **AND** the cron job prompt SHALL direct runtime automation to call `workflow_run` with the selected workflow file path
- **AND** the command output SHALL include the registered cron job id and SHALL NOT claim that schedule registration is not implemented

#### Scenario: Invalid schedule is rejected before registration
- **WHEN** user runs `lango workflow run report.flow.yaml --schedule "not-cron"`
- **THEN** the CLI SHALL reject the schedule before creating a cron job
- **AND** SHALL return an actionable invalid workflow schedule error

### Requirement: Workflow run output routing
`lango workflow run` SHALL write its validation, schedule guidance, and direct execution status output through the Cobra command output stream so wrappers and test harnesses can capture non-error output without intercepting process-global stdout.

#### Scenario: Workflow run schedule registration writes to command output
- **WHEN** user runs `lango workflow run report.flow.yaml --schedule "0 9 * * MON"`
- **THEN** the validated workflow summary and cron registration confirmation write to the Cobra command output stream
