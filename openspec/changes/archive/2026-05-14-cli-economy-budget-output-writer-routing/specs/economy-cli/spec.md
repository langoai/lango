## ADDED Requirements

### Requirement: Budget CLI output routing

`lango economy budget` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture enabled, disabled, and task-guidance output without intercepting process-global stdout.

#### Scenario: Budget enabled output writes to command output
- **WHEN** `lango economy budget` is run with economy enabled
- **THEN** the command writes the budget configuration to the Cobra command output stream

#### Scenario: Budget task guidance writes to command output
- **WHEN** `lango economy budget --task-id=task-1` is run
- **THEN** the command writes the task guidance to the Cobra command output stream

#### Scenario: Budget disabled output writes to command output
- **WHEN** `lango economy budget` is run with economy disabled
- **THEN** the command writes the disabled-state message to the Cobra command output stream
