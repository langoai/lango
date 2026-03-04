## MODIFIED Requirements

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
