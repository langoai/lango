## MODIFIED Requirements

### Requirement: Config delete command
The system SHALL provide a `lango config delete <name>` command with confirmation prompt. When `--force` is not supplied, the confirmation SHALL be driven through the shared confirmation helper using Cobra command input/output streams so wrappers and tests can capture the interaction without replacing process-global stdio.

#### Scenario: Delete with confirmation
- **WHEN** `lango config delete staging` is run without `--force`
- **THEN** a confirmation prompt is shown before deletion

#### Scenario: Delete with force flag
- **WHEN** `lango config delete staging --force` is run
- **THEN** the profile is deleted without confirmation

#### Scenario: Delete denied through command input
- **WHEN** `lango config delete staging` is run and the user answers `n`
- **THEN** the command SHALL print `Aborted.`
- **AND** the profile SHALL remain undeleted
